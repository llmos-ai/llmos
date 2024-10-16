// Package applyinator is a library to apply plans to a host.
// It is referred from: https://github.com/rancher/system-agent/blob/bf4eb096f5a79009badec37f59b16b60c36c191a/pkg/applyinator/applyinator.go
package applyinator

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/llmos-ai/llmos/pkg/applyinator/image"
	"github.com/llmos-ai/llmos/pkg/applyinator/prober"
)

type Applyinator struct {
	mu              *sync.Mutex
	workDir         string
	preserveWorkDir bool
	appliedPlanDir  string
	interlockDir    string
	imageUtil       *image.Utility
}

// CalculatedPlan is passed into Applyinator and is a Plan with checksum calculated
type CalculatedPlan struct {
	Plan     Plan
	Checksum string
}

type Plan struct {
	Files                []File                  `json:"files,omitempty"`
	OneTimeInstructions  []OneTimeInstruction    `json:"instructions,omitempty"`
	Probes               map[string]prober.Probe `json:"probes,omitempty"`
	PeriodicInstructions []PeriodicInstruction   `json:"periodicInstructions,omitempty"`
}

type CommonInstruction struct {
	Name    string   `json:"name,omitempty"`
	Image   string   `json:"image,omitempty"`
	Env     []string `json:"env,omitempty"`
	Args    []string `json:"args,omitempty"`
	Command string   `json:"command,omitempty"`
}

type PeriodicInstruction struct {
	CommonInstruction
	PeriodSeconds    int  `json:"periodSeconds,omitempty"` // default 600, i.e. 10 minutes
	SaveStderrOutput bool `json:"saveStderrOutput,omitempty"`
}

type PeriodicInstructionOutput struct {
	Name                  string `json:"name"`
	Stdout                []byte `json:"stdout"`                // Stdout is a byte array of the gzip+base64 stdout output
	Stderr                []byte `json:"stderr"`                // Stderr is a byte array of the gzip+base64 stderr output
	ExitCode              int    `json:"exitCode"`              // ExitCode is an int representing the exit code of the last run instruction
	LastSuccessfulRunTime string `json:"lastSuccessfulRunTime"` // LastSuccessfulRunTime is a time.UnixDate formatted string of the last successful time (exit code 0) the instruction was run
	Failures              int    `json:"failures"`              // Failures is the number of time the periodic instruction has failed to run
	LastFailedRunTime     string `json:"lastFailedRunTime"`     // LastFailedRunTime is a time.UnixDate formatted string of the time that the periodic instruction started failing
}

type OneTimeInstruction struct {
	CommonInstruction
	SaveOutput bool `json:"saveOutput,omitempty"`
}

// Path would be `/etc/kubernetes/ssl/ca.pem`, Content is base64 encoded.
// If Directory is true, then we are creating a directory, not a file
type File struct {
	Content     string `json:"content,omitempty"`
	Directory   bool   `json:"directory,omitempty"`
	UID         int    `json:"uid,omitempty"`
	GID         int    `json:"gid,omitempty"`
	Path        string `json:"path,omitempty"`
	Permissions string `json:"permissions,omitempty"` // internally, the string will be converted to a uint32 to satisfy os.FileMode
}

const appliedPlanFileSuffix = "-applied.plan"
const applyinatorDateCodeLayout = "20060102-150405"
const defaultCommand = "/run.sh"
const cattleAgentExecutionPwdEnvKey = "CATTLE_AGENT_EXECUTION_PWD"
const cattleAgentAttemptKey = "CATTLE_AGENT_ATTEMPT_NUMBER"
const planRetentionPolicyCount = 64
const restartPendingInterlockFile = "restart-pending"
const applyinatorActiveInterlockFile = "applyinator-active"
const restartPendingTimeout = 5 * time.Minute // Wait a maximum of 5 minutes before force-applying a plan if a restart is pending.

func NewApplyinator(workDir string, preserveWorkDir bool, appliedPlanDir, interlockDir string, imageUtil *image.Utility) *Applyinator {
	return &Applyinator{
		mu:              &sync.Mutex{},
		workDir:         workDir,
		preserveWorkDir: preserveWorkDir,
		appliedPlanDir:  appliedPlanDir,
		interlockDir:    interlockDir,
		imageUtil:       imageUtil,
	}
}

func CalculatePlan(rawPlan []byte) (CalculatedPlan, error) {
	var cp CalculatedPlan
	var plan Plan
	if err := json.Unmarshal(rawPlan, &plan); err != nil {
		return cp, err
	}

	cp.Checksum = checksum(rawPlan)
	cp.Plan = plan

	return cp, nil
}

func checksum(input []byte) string {
	h := sha256.New()
	h.Write(input)

	return fmt.Sprintf("%x", h.Sum(nil))
}

type ApplyOutput struct {
	OneTimeOutput          []byte
	OneTimeApplySucceeded  bool
	PeriodicOutput         []byte
	PeriodicApplySucceeded bool
}

type ApplyInput struct {
	CalculatedPlan             CalculatedPlan
	RunOneTimeInstructions     bool
	OneTimeInstructionAttempts int
	ReconcileFiles             bool
	ExistingOneTimeOutput      []byte
	ExistingPeriodicOutput     []byte
}

// Apply accepts a context, calculated plan, a bool to indicate whether to run the onetime instructions, the existing onetimeinstruction output, and an input byte slice which is a base64+gzip json-marshalled map of PeriodicInstructionOutput
// entries where the key is the PeriodicInstructionOutput.Name. It outputs a revised versions of the existing outputs, and if specified, runs the one time instructions. Notably, ApplyOutput.OneTimeApplySucceeded will be false if ApplyInput.RunOneTimeInstructions is false
func (a *Applyinator) Apply(ctx context.Context, input ApplyInput) (ApplyOutput, error) {
	logrus.Debugf("[Applyinator] Applying plan with checksum %s", input.CalculatedPlan.Checksum)
	logrus.Tracef("[Applyinator] Applying plan - attempting to get lock")
	output := ApplyOutput{
		OneTimeOutput:  input.ExistingOneTimeOutput,
		PeriodicOutput: input.ExistingPeriodicOutput,
	}
	a.mu.Lock()
	logrus.Tracef("[Applyinator] Applying plan - lock achieved")
	defer a.mu.Unlock()
	now := time.Now()
	nowUnixTimeString := now.Format(time.UnixDate)
	nowString := now.Format(applyinatorDateCodeLayout)

	// Check to see if we are safe to apply.
	if a.interlockDir != "" {
		restartPendingInterlockFilePath := filepath.Join(a.interlockDir, restartPendingInterlockFile)
		applyinatorActiveInterlockFilePath := filepath.Join(a.interlockDir, applyinatorActiveInterlockFile)
		// First off, remove check and remove the active interlock as the applyinator is not actually active
		if _, err := os.Stat(applyinatorActiveInterlockFile); err == nil {
			err = os.Remove(applyinatorActiveInterlockFile)
			if err != nil {
				logrus.Errorf("unable to remove applyinator active interlock file %s: %v", applyinatorActiveInterlockFilePath, err)
			}
		}

		if _, err := os.Stat(restartPendingInterlockFilePath); err == nil {
			// check the restart pending interlock file to see if we've passed our threshold for blocking
			fileContents, err := os.ReadFile(restartPendingInterlockFilePath)
			if err != nil {
				return output, fmt.Errorf("unable to read restart pending interlock file %s: %w", restartPendingInterlockFilePath, err)
			}
			// Parse the time out of the file and determine if we have passed our time threshold
			t, err := time.Parse(time.UnixDate, string(fileContents))
			if err != nil {
				// If we are unable to parse the first observed time out of the file, write "now" as the first observed time of the file.
				if err := os.WriteFile(restartPendingInterlockFilePath, []byte(nowUnixTimeString), 0600); err != nil {
					return output, fmt.Errorf("unable to write first-observed time to restart pending interlock file %s: %w", restartPendingInterlockFilePath, err)
				}
				return output, fmt.Errorf("restart is pending for system-agent, waiting %s until ignoring pending restart", restartPendingTimeout.String())
			}
			if now.Before(t.Add(restartPendingTimeout)) {
				return output, fmt.Errorf("restart is pending for system-agent, waiting %s until ignoring pending restart", t.Add(restartPendingTimeout).Sub(now).String())
			}
			// remove the restart pending file
			err = os.Remove(restartPendingInterlockFilePath)
			if err != nil {
				logrus.Errorf("error encountered while removing restart pending interlock file %s: %v", restartPendingInterlockFilePath, err)
			}
		}

		// At this point, there is no restart-pending and we can continue with applyinator reconciliation, so create the applyinator-active file
		err := os.WriteFile(applyinatorActiveInterlockFilePath, []byte(nowUnixTimeString), 0600)
		if err != nil {
			logrus.Errorf("unable to write applyinator active interlock file %s: %v", applyinatorActiveInterlockFilePath, err)
		}
		defer func() {
			// Remove the Applyinator Active Interlock File
			err = os.Remove(applyinatorActiveInterlockFilePath)
			if err != nil {
				logrus.Errorf("unable to remove applyinator active interlock file %s: %v", applyinatorActiveInterlockFilePath, err)
			}
		}()
	}

	executionDir := filepath.Join(a.workDir, nowString)
	logrus.Tracef("[Applyinator] Applying calculated node plan contents %v", input.CalculatedPlan.Checksum)
	logrus.Tracef("[Applyinator] Using %s as execution directory", executionDir)
	if a.appliedPlanDir != "" {
		logrus.Debugf("[Applyinator] Writing applied calculated plan contents to historical plan directory %s", a.appliedPlanDir)
		if err := os.MkdirAll(a.appliedPlanDir, 0700); err != nil {
			logrus.Errorf("error creawting applied plan directory: %v", err)
		}
		if err := a.writePlanToDisk(now, &input.CalculatedPlan); err != nil {
			logrus.Errorf("error writing applied plan to disk: %v", err)
		}
		if err := a.appliedPlanRetentionPolicy(planRetentionPolicyCount); err != nil {
			logrus.Errorf("error while applying plan retention policy: %v", err)
		}
	}

	if input.ReconcileFiles {
		for _, file := range input.CalculatedPlan.Plan.Files {
			if file.Directory {
				logrus.Debugf("[Applyinator] Creating directory %s", file.Path)
				if err := CreateDirectory(file); err != nil {
					return output, err
				}
			} else {
				logrus.Debugf("[Applyinator] Writing file %s", file.Path)
				if err := writeBase64ContentToFile(file); err != nil {
					return output, err
				}
			}
		}
	}

	if !a.preserveWorkDir {
		logrus.Debugf("[Applyinator] Cleaning working directory before applying %s", a.workDir)
		if err := os.RemoveAll(a.workDir); err != nil {
			return output, err
		}
	}
	if input.RunOneTimeInstructions {
		logrus.Infof("[Applyinator] Applying one-time instructions for plan with checksum %s", input.CalculatedPlan.Checksum)
		executionOutputs := map[string][]byte{}
		if len(input.ExistingOneTimeOutput) > 0 {
			objectBuffer, err := generateByteBufferFromBytes(input.ExistingOneTimeOutput)
			if err != nil {
				return output, err
			}
			if err := json.Unmarshal(objectBuffer.Bytes(), &executionOutputs); err != nil {
				return output, err
			}
		}

		oneTimeApplySucceeded := true
		for index, instruction := range input.CalculatedPlan.Plan.OneTimeInstructions {
			logrus.Debugf("[Applyinator] Executing instruction %d attempt %d for plan %s", index, input.OneTimeInstructionAttempts, input.CalculatedPlan.Checksum)
			executionInstructionDir := filepath.Join(executionDir, input.CalculatedPlan.Checksum+"_"+strconv.Itoa(index))
			prefix := input.CalculatedPlan.Checksum + "_" + strconv.Itoa(index)
			executeOutput, _, exitCode, err := a.execute(ctx, prefix, executionInstructionDir, instruction.CommonInstruction, true, input.OneTimeInstructionAttempts)
			if err != nil || exitCode != 0 {
				logrus.Errorf("error executing instruction %d %s: %v", index, instruction.Name, err)
				oneTimeApplySucceeded = false
			}
			if instruction.Name == "" && instruction.SaveOutput {
				logrus.Errorf("instruction does not have a name set, cannot save output data")
			} else if instruction.SaveOutput {
				executionOutputs[instruction.Name] = executeOutput
			}
			// If we have failed to apply our one-time instructions, we need to break in order to stop subsequent instructions from executing.
			if !oneTimeApplySucceeded {
				break
			}
		}

		output.OneTimeApplySucceeded = oneTimeApplySucceeded

		marshalledExecutionOutputs, err := json.Marshal(executionOutputs)
		if err != nil {
			return output, err
		}

		oneTimeApplyOutput, err := gzipByteSlice(marshalledExecutionOutputs)
		if err != nil {
			return output, err
		}

		output.OneTimeOutput = oneTimeApplyOutput
	}

	periodicOutputs := map[string]PeriodicInstructionOutput{}
	if len(input.ExistingPeriodicOutput) > 0 {
		objectBuffer, err := generateByteBufferFromBytes(input.ExistingPeriodicOutput)
		if err != nil {
			return output, err
		}
		if err := json.Unmarshal(objectBuffer.Bytes(), &periodicOutputs); err != nil {
			return output, err
		}
	}

	periodicApplySucceeded := true
	for index, instruction := range input.CalculatedPlan.Plan.PeriodicInstructions {
		if instruction.Name == "" {
			logrus.Errorf("periodic instruction %d did not have name, unable to run", index)
			continue
		}
		var previousRunTime, lastFailureTime string
		var failures int
		if po, ok := periodicOutputs[instruction.Name]; ok {
			if po.LastSuccessfulRunTime != "" {
				logrus.Debugf("[Applyinator] Got periodic output for instruction %s and am now parsing last successful run time %s", instruction.Name, po.LastSuccessfulRunTime)
				t, err := time.Parse(time.UnixDate, po.LastSuccessfulRunTime)
				if err != nil {
					logrus.Errorf("error encountered during parsing of last successful run time: %v", err)
				} else {
					previousRunTime = po.LastSuccessfulRunTime
					if instruction.PeriodSeconds == 0 {
						instruction.PeriodSeconds = 600 // set default period to 600 seconds
					}
					if now.Before(t.Add(time.Second*time.Duration(instruction.PeriodSeconds))) && !input.RunOneTimeInstructions {
						logrus.Debugf("[Applyinator] Not running periodic instruction %s as period duration has not elapsed since last successful run", instruction.Name)
						continue
					}
				}
			}
			if po.LastFailedRunTime != "" {
				logrus.Debugf("[Applyinator] Got periodic output for instruction %s and am now parsing last failed time %s", instruction.Name, po.LastFailedRunTime)
				t, err := time.Parse(time.UnixDate, po.LastFailedRunTime)
				if err != nil {
					logrus.Errorf("error encountered during parsing of last failed run time: %+v", err)
				} else {
					lastFailureTime = po.LastFailedRunTime
					failures = po.Failures
					failureCooldown := failures
					if failures > 6 {
						failureCooldown = 6
					} else if failures == 0 {
						failureCooldown = 1
					}
					logrus.Debugf("[Applyinator] Instruction %s - Last failed run attempt was %s, failures: %d, failureCooldown: %d", instruction.Name, lastFailureTime, failures, failureCooldown)
					if now.Before(t.Add(time.Second*time.Duration(30*failureCooldown))) && !input.RunOneTimeInstructions {
						logrus.Debugf("[Applyinator] Not running periodic instruction %s as failure cooldown has not elapsed since last failed run", instruction.Name)
						continue
					}
				}
			}
		}
		logrus.Debugf("[Applyinator] Executing periodic instruction %d for plan %s", index, input.CalculatedPlan.Checksum)
		executionInstructionDir := filepath.Join(executionDir, input.CalculatedPlan.Checksum+"_"+strconv.Itoa(index))
		prefix := input.CalculatedPlan.Checksum + "_" + strconv.Itoa(index)
		stdout, stderr, exitCode, err := a.execute(ctx, prefix, executionInstructionDir, instruction.CommonInstruction, false, failures+1)
		if err != nil || exitCode != 0 {
			periodicApplySucceeded = false
		}
		lsrt := nowUnixTimeString
		if exitCode != 0 {
			lsrt = previousRunTime
			lastFailureTime = nowUnixTimeString
			failures++
		} else {
			// reset last failure time and failure count
			lastFailureTime = ""
			failures = 0
		}
		if !instruction.SaveStderrOutput {
			stderr = []byte{}
		}
		periodicOutputs[instruction.Name] = PeriodicInstructionOutput{
			Name:                  instruction.Name,
			Stdout:                stdout,
			Stderr:                stderr,
			ExitCode:              exitCode,
			LastSuccessfulRunTime: lsrt,
			LastFailedRunTime:     lastFailureTime,
			Failures:              failures,
		}
		if !periodicApplySucceeded {
			break
		}
	}

	output.PeriodicApplySucceeded = periodicApplySucceeded

	marshalledExecutionOutputs, err := json.Marshal(periodicOutputs)
	if err != nil {
		return output, err
	}
	periodicApplyOutput, err := gzipByteSlice(marshalledExecutionOutputs)
	if err != nil {
		return output, err
	}

	output.PeriodicOutput = periodicApplyOutput
	return output, nil
}

func gzipByteSlice(input []byte) ([]byte, error) {
	var gzOutput bytes.Buffer

	gzWriter := gzip.NewWriter(&gzOutput)

	_, err := gzWriter.Write(input)
	if err != nil {
		return []byte{}, err
	}

	if err := gzWriter.Close(); err != nil {
		return []byte{}, err
	}
	return gzOutput.Bytes(), nil
}

func generateByteBufferFromBytes(input []byte) (*bytes.Buffer, error) {
	buffer := bytes.NewBuffer(input)
	gzReader, err := gzip.NewReader(buffer)
	if err != nil {
		return nil, err
	}

	var objectBuffer bytes.Buffer
	_, err = io.Copy(&objectBuffer, gzReader)
	if err != nil {
		return nil, err
	}
	return &objectBuffer, nil
}

func (a *Applyinator) appliedPlanRetentionPolicy(retention int) error {
	planFiles, err := a.getAppliedPlanFiles()
	if err != nil {
		return err
	}

	if len(planFiles) <= retention {
		return nil
	}

	sort.Slice(planFiles, func(i, j int) bool {
		return planFiles[i].Name() < planFiles[j].Name()
	})

	delCount := len(planFiles) - retention
	for _, df := range planFiles[:delCount] {
		historicalPlanFile := filepath.Join(a.appliedPlanDir, df.Name())
		logrus.Infof("[Applyinator] Removing historical applied plan (retention policy count: %d) %s", retention, historicalPlanFile)
		if err := os.Remove(historicalPlanFile); err != nil {
			return err
		}
	}
	return nil
}

func (a *Applyinator) getAppliedPlanFiles() ([]os.DirEntry, error) {
	var planFiles []os.DirEntry
	dirListedPlanFiles, err := os.ReadDir(a.appliedPlanDir)
	if err != nil {
		return nil, err
	}

	for _, f := range dirListedPlanFiles {
		if strings.HasSuffix(f.Name(), appliedPlanFileSuffix) && !f.IsDir() {
			planFiles = append(planFiles, f)
		}
	}
	return planFiles, nil
}

func (a *Applyinator) writePlanToDisk(now time.Time, plan *CalculatedPlan) error {
	planFiles, err := a.getAppliedPlanFiles()
	if err != nil {
		return err
	}

	file := now.Format(applyinatorDateCodeLayout) + appliedPlanFileSuffix
	anpString, err := json.Marshal(plan)
	if err != nil {
		return err
	}

	if len(planFiles) != 0 {
		sort.Slice(planFiles, func(i, j int) bool {
			return planFiles[i].Name() > planFiles[j].Name()
		})
		existingFileContent, err := os.ReadFile(filepath.Join(a.appliedPlanDir, planFiles[0].Name()))
		if err != nil {
			return err
		}
		if bytes.Equal(existingFileContent, anpString) {
			logrus.Debugf("[Applyinator] Not writing applied plan to file %s as the last file written (%s) had identical contents", file, planFiles[0].Name())
			return nil
		}
	}

	return writeContentToFile(filepath.Join(a.appliedPlanDir, file), os.Getuid(), os.Getgid(), 0600, anpString)
}

func (a *Applyinator) execute(ctx context.Context, prefix, executionDir string, instruction CommonInstruction, combinedOutput bool, attempt int) ([]byte, []byte, int, error) {
	if instruction.Image == "" {
		logrus.Infof("[Applyinator] No image provided, creating empty working directory %s", executionDir)
		if err := CreateDirectory(File{Directory: true, Path: executionDir}); err != nil {
			logrus.Errorf("error while creating empty working directory: %v", err)
			return nil, nil, -1, err
		}
	} else {
		logrus.Infof("[Applyinator] Extracting image %s to directory %s", instruction.Image, executionDir)
		if err := a.imageUtil.Stage(executionDir, instruction.Image); err != nil {
			logrus.Errorf("error while staging: %v", err)
			return nil, nil, -1, err
		}
	}

	command := instruction.Command

	if command == "" {
		logrus.Debugf("[Applyinator] Command was not specified, defaulting to %s%s", executionDir, defaultCommand)
		command = executionDir + defaultCommand
	}

	cmd := exec.CommandContext(ctx, command, instruction.Args...)
	logrus.Infof("[Applyinator] Running command: %s %v", instruction.Command, instruction.Args)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, instruction.Env...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", cattleAgentExecutionPwdEnvKey, executionDir))
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%d", cattleAgentAttemptKey, attempt))
	cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH")+":"+executionDir)
	cmd.Dir = executionDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logrus.Errorf("error setting up stdout pipe: %v", err)
		return nil, nil, -1, err
	}
	defer stdout.Close() // nolint: errcheck

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logrus.Errorf("error setting up stderr pipe: %v", err)
		return nil, nil, -1, err
	}
	defer stderr.Close() // nolint: errcheck

	var (
		eg              = errgroup.Group{}
		stdoutWriteLock *sync.Mutex
		stderrWriteLock *sync.Mutex
		stdoutBuffer    bytes.Buffer
		stderrBuffer    bytes.Buffer
	)

	if combinedOutput {
		stderrBuffer = stdoutBuffer
		stdoutWriteLock = &sync.Mutex{}
		stderrWriteLock = stdoutWriteLock
	} else {
		stdoutWriteLock = &sync.Mutex{}
		stderrWriteLock = &sync.Mutex{}
	}

	eg.Go(func() error {
		return streamLogs("["+prefix+":stdout]", &stdoutBuffer, stdout, stdoutWriteLock)
	})
	eg.Go(func() error {
		return streamLogs("["+prefix+":stderr]", &stderrBuffer, stderr, stderrWriteLock)
	})

	if err := cmd.Start(); err != nil {
		return nil, nil, -1, err
	}

	// Wait for I/O to complete before calling cmd.Wait() because cmd.Wait() will close the I/O pipes.
	_ = eg.Wait()
	exitCode := 0
	if err := cmd.Wait(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
		}
	}
	logrus.Infof("[Applyinator] Command %s %v finished with err: %v and exit code: %d", instruction.Command, instruction.Args, err, exitCode)
	return stdoutBuffer.Bytes(), stderrBuffer.Bytes(), exitCode, err
}

// streamLogs accepts a prefix, outputBuffer, reader, and buffer lock and will scan input from the reader and write it
// to the output buffer while also logging anything that comes from the reader with the prefix.
func streamLogs(prefix string, outputBuffer *bytes.Buffer, reader io.Reader, lock *sync.Mutex) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		logrus.Infof("%s: %s", prefix, scanner.Text())
		lock.Lock()
		outputBuffer.Write(append(scanner.Bytes(), []byte("\n")...))
		lock.Unlock()
	}
	return nil
}
