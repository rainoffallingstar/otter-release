rule qcsummary:
  message:"summary QC in the pipeline ..."
  input:
    expand(os.path.join(config["output"]["trim_dir"], "{sample}_R1.fastq.gz_trimming_report.txt"),sample = config["SIDs"]),
    expand(os.path.join(config["output"]["trim_dir"], "{sample}_R2.fastq.gz_trimming_report.txt"),sample = config["SIDs"]),
    expand(os.path.join(config["directories"]["qc"]["after"], "{sample}_val_1_fastqc.zip"), sample=config["SIDs"]),
    expand(os.path.join(config["directories"]["qc"]["after"], "{sample}_val_2_fastqc.zip"), sample=config["SIDs"]),
    expand(os.path.join(config["directories"]["qc"]["after"], "{sample}_val_1_fastqc.html"), sample=config["SIDs"]),
    expand(os.path.join(config["directories"]["qc"]["after"], "{sample}_val_2_fastqc.html"), sample=config["SIDs"]),
    expand(os.path.join(config["directories"]["qc"]["main"], "{sample}_seqkit_stat.txt"), sample=config["SIDs"]) ,
    expand(os.path.join(config["directories"]["bsmap"]["main"], "{sample}_{species}.bam"), sample=config["SIDs"],species =config["species"]) ,
    expand(os.path.join(config["directories"]["qualimap"],"{sample}_{species}","qualimapReport.html") , sample=config["SIDs"],species =config["species"]),
    os.path.join(config["directories"]["methylation_call"], "methrixh5","CpG_coverage.xlsx")
  output:
    os.path.join(config["directories"]["qc_summary"],"qc_summary.txt")
  params:
    self_config = config["selfconfig"],
    qc_output = config["directories"]["qc_summary"]
    
  threads:5
  shell:
    """
    enva run base Rscript R/QC_summary.R --configfolder {params.self_config} --output_dir {params.qc_output}

    """
    
