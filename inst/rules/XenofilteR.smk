rule xenofilteR:
  message:"xenofilteR ..."
  input:
    expand(os.path.join(config["directories"]["bsmap"]["main"], "{sample}_{species}.bam"), sample=config["metadata"]["sample_ids"],species = config["workflow"]["species"]["name"])
  output:
    os.path.join(config["directories"]["bsmap"]["main"],"Filtered_bams" ,"filtered_success.txt")
  params:
    filter_root = config["directories"]["bsmap"]["main"],
    host = config["workflow"]["species"]["host"],
    graft = config["workflow"]["species"]["graft"],
    MM_threshold = (4 if config["mode"] == "RNASEQ" else 6),
    Unmapped_penalty = 8,
    mode = config["mode"],
    graft_bams = [os.path.join(config["directories"]["bsmap"]["main"], f"{s}_{config['workflow']['species']['graft']}.bam") for s in config["metadata"]["sample_ids"]],
    host_bams = [os.path.join(config["directories"]["bsmap"]["main"], f"{s}_{config['workflow']['species']['host']}.bam") for s in config["metadata"]["sample_ids"]],
    graft_ref = config["reference"]["files"]["fasta"][config["workflow"]["species"]["name"].index(config["workflow"]["species"]["graft"])] if config["workflow"]["species"]["host"] else "",
    host_ref = config["reference"]["files"]["fasta"][config["workflow"]["species"]["name"].index(config["workflow"]["species"]["host"])] if config["workflow"]["species"]["host"] else ""
  threads:4
  shell:
    """
    mkdir -p {params.filter_root}/Filtered_bams

    xenofilter run \
      --graft {params.graft_bams} \
      --host {params.host_bams} \
      --output {params.filter_root}/Filtered_bams \
      --mm-threshold {params.MM_threshold} \
      --unmapped-penalty {params.Unmapped_penalty} \
      --threads {threads} \
      --recalculate-nm \
      {"--graft-ref " + params.graft_ref if params.graft_ref else ""}\
      {"--host-ref " + params.host_ref if params.host_ref else ""}\
      {" --bisulfite" if params.mode != "RNASEQ" else ""} \
      && touch {params.filter_root}/Filtered_bams/filtered_success.txt

    """
    
