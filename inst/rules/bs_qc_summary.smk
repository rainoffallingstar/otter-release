rule qcsummary:
  message:"summary QC in the pipeline ..."
  input:
    expand(os.path.join(config["output"]["trim_dir"], "{sample}_R1.fastq.gz_trimming_report.txt"),sample = config["metadata"]["sample_ids"]),
    expand(os.path.join(config["output"]["trim_dir"], "{sample}_R2.fastq.gz_trimming_report.txt"),sample = config["metadata"]["sample_ids"]),
    expand(os.path.join(config["directories"]["qc"]["before"], "{sample}_R1_fqc", "fastqc_data.txt"), sample=config["metadata"]["sample_ids"]),
    expand(os.path.join(config["directories"]["qc"]["before"], "{sample}_R2_fqc", "fastqc_data.txt"), sample=config["metadata"]["sample_ids"]),
    expand(os.path.join(config["directories"]["qc"]["after"], "{sample}_val_1_fqc", "fastqc_data.txt"), sample=config["metadata"]["sample_ids"]),
    expand(os.path.join(config["directories"]["qc"]["after"], "{sample}_val_2_fqc", "fastqc_data.txt"), sample=config["metadata"]["sample_ids"]),
    expand(os.path.join(config["directories"]["bsmap"]["main"], "{sample}_{species}.bam"), sample=config["metadata"]["sample_ids"],species =config["workflow"]["species"]["name"]) ,
    expand(os.path.join(config["directories"]["qualimap"],"{sample}_{species}","qualimapReport.html") , sample=config["metadata"]["sample_ids"],species =config["workflow"]["species"]["name"]),
    os.path.join(config["directories"]["methylation_call"], "methrixh5","CpG_coverage.xlsx")
  output:
    os.path.join(config["directories"]["qc_summary"],"qc_summary.xlsx")
  params:
    self_config = config["directories"]["selfconfig"],
    qc_output = config["directories"]["qc_summary"]
    
  threads:5
  shell:
    """
    qctb --config {params.self_config}/config.yaml --output {params.qc_output}/qc_summary.xlsx
    """
    
