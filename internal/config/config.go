package config

import (
	"time"
)

// WorkflowConfig represents the main workflow configuration
type WorkflowConfig struct {
	// Basic configuration
	Mode     string `mapstructure:"mode"` // RRBS/WGBS/RNASEQ/PDX
	Species1 string `mapstructure:"species1"`
	Species2 string `mapstructure:"species2,omitempty"`

	// FASTQ configuration
	Suffix1 string `mapstructure:"suffix1"` // default: "_R1.fastq.gz"
	Suffix2 string `mapstructure:"suffix2"` // default: "_R2.fastq.gz" (auto-derived)

	// Input configuration
	Input InputConfig `mapstructure:"input"`

	// Output configuration
	Output OutputConfig `mapstructure:"output"`

	// Reference configuration
	Reference ReferenceConfig `mapstructure:"reference"`

	// Engine configuration
	Engine EngineConfig `mapstructure:"engine"`

	// Parallel processing
	Parallel ParallelConfig `mapstructure:"parallel"`

	// Advanced parameters
	Adapters  map[string]string `mapstructure:"adapters"`
	Alignment AlignmentConfig   `mapstructure:"alignment"`

	// Workflow control
	WorkflowDir string `mapstructure:"workflow_dir"`
	AnalysisDir string `mapstructure:"analysis_dir"`
	SelfConfig  string `mapstructure:"self_config"`
	QCDir       string `mapstructure:"qc_dir"`
	TrimDir     string `mapstructure:"trim_dir"`

	// Metadata
	CreatedAt time.Time `mapstructure:"-"`
	JobID     string    `mapstructure:"job_id"`
	UserID    string    `mapstructure:"user_id"`
	UserEmail string    `mapstructure:"user_email"`
}

// InputConfig represents input file configuration
type InputConfig struct {
	FastqDir  string `mapstructure:"fastq_dir"`
	PdataFile string `mapstructure:"pdata_file"`
}

// OutputConfig represents output directory configuration
type OutputConfig struct {
	BaseDir     string `mapstructure:"base_dir"`
	WorkflowDir string `mapstructure:"workflow_dir"`
	AnalysisDir string `mapstructure:"analysis_dir"`
	QCDir       string `mapstructure:"qc_dir"`
	TrimDir     string `mapstructure:"trim_dir"`
	OutDirMCall string `mapstructure:"out_dir_mcall"`
	OutDirUmx   string `mapstructure:"out_dir_umx"`
	OutDirBetaM string `mapstructure:"out_dir_beta_m"`
	LogDir      string `mapstructure:"log_dir"`
}

// ReferenceConfig represents reference genome configuration
type ReferenceConfig struct {
	Genome       string            `mapstructure:"genome"`
	GenomeFasta  []string          `mapstructure:"genome_fasta"`
	GenomeIndex  []string          `mapstructure:"genome_index"`
	GenomeAnno   []string          `mapstructure:"genome_anno"`
	RNASEQGTF    string            `mapstructure:"rnaseq_gtf"`
	RNASEQRef    string            `mapstructure:"rnaseq_ref"`
	Indices      map[string]string `mapstructure:"indices"`
	AutoValidate bool              `mapstructure:"auto_validate"`
}

// EngineConfig represents execution engine configuration
type EngineConfig struct {
	Type  string      `mapstructure:"type"` // auto/slurm/local
	Slurm SlurmConfig `mapstructure:"slurm,omitempty"`
	Local LocalConfig `mapstructure:"local,omitempty"`
}

// SlurmConfig represents Slurm cluster configuration
type SlurmConfig struct {
	Partition  string `mapstructure:"partition"`
	Cores      int    `mapstructure:"cores"`
	Memory     string `mapstructure:"memory"`
	JobName    string `mapstructure:"job_name"`
	MaxRetries int    `mapstructure:"max_retries"`
}

// LocalConfig represents local execution configuration
type LocalConfig struct {
	MaxCores  int    `mapstructure:"max_cores"`
	MaxMemory string `mapstructure:"max_memory"`
}

// ParallelConfig represents parallel processing configuration
type ParallelConfig struct {
	Workers int `mapstructure:"workers"`
}

// AlignmentConfig represents alignment parameters
type AlignmentConfig struct {
	C1        string  `mapstructure:"C1"`         // e.g., "7"
	C2        string  `mapstructure:"C2"`         // e.g., "9"
	T1        int     `mapstructure:"T1"`         // e.g., 0
	T2        int     `mapstructure:"T2"`         // e.g., 0
	ErrorRate float64 `mapstructure:"error_rate"` // e.g., 0.2

	// Trim parameters
	Read1_5 int `mapstructure:"read1_5"`  // e.g., 0
	Read1_3 int `mapstructure:"read1_3"`  // e.g., 0
	Read2_5 int `mapstructure:"read2_5"`  // e.g., 0
	Read2_3 int `mapstructure:"read2_3"`  // e.g., 0
	SeqDeth int `mapstructure:"seq_deth"` // e.g., 10
}
