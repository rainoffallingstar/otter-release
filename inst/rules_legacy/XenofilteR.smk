rule xenofilteR:
  message:"xenofilteR ..."
  input:
    expand(os.path.join(config["directories"]["bsmap"]["main"], "{sample}_{species}_pdx_patch_success"),sample=config["metadata"]["sample_ids"],species = config["workflow"]["species"]["name"]),
    expand(os.path.join(config["directories"]["bsmap"]["main"], "{sample}_fixed_{species}.bam"), sample=config["metadata"]["sample_ids"],species = config["workflow"]["species"]["name"])
  output:
    os.path.join(config["directories"]["bsmap"]["main"],"Filtered_bams" ,"filtered_success.txt")
  params:
    filter_root = config["directories"]["bsmap"]["main"],
    host = config["workflow"]["species"]["host"],
    graft = config["workflow"]["species"]["graft"],
    MM_threshold = (4 if config["mode"] == "RNASEQ" else 6),
    Unmapped_penalty = 8,
    mode = config["mode"]
  threads:4
  shell:
    """
    enva run base Rscript R/xenofilteR.R -d {params.filter_root} \
    --graft {params.graft} \
    --host {params.host} \
    --threads {threads} \
    --MM_threshold {params.MM_threshold} \
    --Unmapped_penalty {params.Unmapped_penalty} \
    --Mode {params.mode} \
    && touch {params.filter_root}/Filtered_bams/filtered_success.txt
    
    """
    
