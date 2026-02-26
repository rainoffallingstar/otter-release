rule fqcAtclean:
  message:"Running fqc after Trim ..."
  input:
    R1 = os.path.join(config["output"]["trim_dir"], "{sample}_val_1.fq.gz"),
    R2 = os.path.join(config["output"]["trim_dir"], "{sample}_val_2.fq.gz")
  output:
    R1_data = os.path.join(config["directories"]["qc"]["after"], "{sample}_val_1_fqc", "fastqc_data.txt"),
    R2_data = os.path.join(config["directories"]["qc"]["after"], "{sample}_val_2_fqc", "fastqc_data.txt")
  params:
    R1_dir = os.path.join(config["directories"]["qc"]["after"], "{sample}_val_1_fqc"),
    R2_dir = os.path.join(config["directories"]["qc"]["after"], "{sample}_val_2_fqc")
  threads: 2
  shell:
    """
    fqc -q {input.R1} -s {params.R1_dir} --no-html
    fqc -q {input.R2} -s {params.R2_dir} --no-html
    """
