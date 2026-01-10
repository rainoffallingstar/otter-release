package config

import (
	"os"
	"time"
)

// LoadDefaults loads default configuration values
func LoadDefaults() *WorkflowConfig {
	// Get user info
	userID := os.Getenv("USER")
	if userID == "" {
		userID = "unknown"
	}

	userEmail := os.Getenv("USER") + "@unknown"
	if os.Getenv("EMAIL") != "" {
		userEmail = os.Getenv("EMAIL")
	}

	return &WorkflowConfig{
		Mode:     "RRBS",
		Species1: "human",
		Species2: "",
		Suffix1:  "_R1.fastq.gz",
		Suffix2:  "", // Will be auto-derived

		Input: InputConfig{
			FastqDir:  "./data/fastq",
			PdataFile: "",
		},

		Output: OutputConfig{
			BaseDir:     "./results",
			WorkflowDir: "/workflow",
			AnalysisDir: "/analysis",
			QCDir:       "/QC",
			TrimDir:     "/trim",
			OutDirMCall: "/mCall",
			OutDirUmx:   "/uxm",
			OutDirBetaM: "/betaM",
			LogDir:      "/log",
		},

		Reference: ReferenceConfig{
			Genome:       "hg19",
			GenomeFasta:  []string{},
			GenomeIndex:  []string{},
			GenomeAnno:   []string{},
			RNASEQGTF:    "",
			RNASEQRef:    "",
			Indices:      map[string]string{},
			AutoValidate: true,
		},

		Engine: EngineConfig{
			Type: "auto",
			Slurm: SlurmConfig{
				Partition:  "amd_512",
				Cores:      20,
				Memory:     "100G",
				JobName:    "xdxtools",
				MaxRetries: 3,
			},
			Local: LocalConfig{
				MaxCores:  8,
				MaxMemory: "32G",
			},
		},

		Parallel: ParallelConfig{
			Workers: 4,
		},

		Adapters: map[string]string{
			"adapter1": "AGATCGGAAGAGC",
			"adapter2": "AGATCGGAAGAGC",
		},

		Alignment: AlignmentConfig{
			C1:        "7",
			C2:        "9",
			T1:        0,
			T2:        0,
			ErrorRate: 0.2,
			Read1_5:   0,
			Read1_3:   0,
			Read2_5:   0,
			Read2_3:   0,
			SeqDeth:   10,
		},

		WorkflowDir: "/workflow",
		AnalysisDir: "/analysis",
		SelfConfig:  "/config",
		QCDir:       "/QC",
		TrimDir:     "/trim",

		CreatedAt: time.Now(),
		JobID:     userID,
		UserID:    userID,
		UserEmail: userEmail,
	}
}
