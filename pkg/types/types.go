package types

import "time"

// Mode represents the workflow mode
type Mode string

const (
	ModeRRBS   Mode = "RRBS"
	ModeWGBS   Mode = "WGBS"
	ModeBSSEQ  Mode = "BSSEQ"
	ModeRNASEQ Mode = "RNASEQ"
)

// EngineType represents the execution engine type
type EngineType string

const (
	EngineSlurm EngineType = "slurm"
	EngineLocal EngineType = "local"
)

// Sample represents a FASTQ sample
type Sample struct {
	Name     string            `json:"name"`
	FastqR1  string            `json:"fastq_r1"`
	FastqR2  string            `json:"fastq_r2,omitempty"`
	Metadata map[string]string `json:"metadata"`
}

// PairedSample represents a paired-end sample
type PairedSample struct {
	Name   string `json:"name"`
	R1Path string `json:"r1_path"`
	R2Path string `json:"r2_path"`
	Valid  bool   `json:"valid"`
}

// PData represents phenotype data
type PData struct {
	Samples []string                     `json:"samples"`
	Columns []string                     `json:"columns"`
	Data    map[string]map[string]string `json:"data"`
}

// Status represents execution status
type Status struct {
	State     string    `json:"state"`
	JobID     string    `json:"job_id"`
	Progress  int       `json:"progress"`
	Message   string    `json:"message"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

// Workflow represents a workflow instance
type Workflow struct {
	ID        string      `json:"id"`
	Config    interface{} `json:"config"` // Use interface{} to avoid circular dependency
	Status    *Status     `json:"status"`
	Samples   []Sample    `json:"samples"`
	Steps     int         `json:"steps"`
	OutputDir string      `json:"output_dir"`
	LogFile   string      `json:"log_file"`
}

// ErrorType represents error categories
type ErrorType string

const (
	ErrorConfig   ErrorType = "CONFIG_ERROR"
	ErrorInput    ErrorType = "INPUT_ERROR"
	ErrorEngine   ErrorType = "ENGINE_ERROR"
	ErrorWorkflow ErrorType = "WORKFLOW_ERROR"
	ErrorScript   ErrorType = "SCRIPT_ERROR"
)

// XDXError represents a custom error
type XDXError struct {
	Type    ErrorType `json:"type"`
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Cause   error     `json:"cause,omitempty"`
}

func (e *XDXError) Error() string {
	return e.Message
}
