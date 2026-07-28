rule xenofilteR:
  message: "Filter PDX graft alignments"
  input:
    graft_bams=expand(
      os.path.join(config["directories"]["bsmap"]["main"], "{sample}_fixed_" + config["workflow"]["species"]["graft"] + ".bam"),
      sample=config["metadata"]["sample_ids"],
    ),
    host_bams=expand(
      os.path.join(config["directories"]["bsmap"]["main"], "{sample}_fixed_" + config["workflow"]["species"]["host"] + ".bam"),
      sample=config["metadata"]["sample_ids"],
    )
  output:
    filtered_bams=expand(
      os.path.join(config["directories"]["bsmap"]["main"], "Filtered_bams", "{sample}_fixed_" + config["workflow"]["species"]["graft"] + "_Filtered.bam"),
      sample=config["metadata"]["sample_ids"],
    ),
    filtered_bais=expand(
      os.path.join(config["directories"]["bsmap"]["main"], "Filtered_bams", "{sample}_fixed_" + config["workflow"]["species"]["graft"] + "_Filtered.bam.bai"),
      sample=config["metadata"]["sample_ids"],
    )
  params:
    filter_root=config["directories"]["bsmap"]["main"],
    host=config["workflow"]["species"]["host"],
    graft=config["workflow"]["species"]["graft"],
    mm_threshold=(4 if config["mode"] == "RNASEQ" else 6),
    unmapped_penalty=8,
    mode=config["mode"],
    graft_ref=config["reference"]["files"]["fasta"][config["workflow"]["species"]["name"].index(config["workflow"]["species"]["graft"])],
    host_ref=config["reference"]["files"]["fasta"][config["workflow"]["species"]["name"].index(config["workflow"]["species"]["host"])]
  threads: 4
  shell:
    """
    set -euo pipefail
    mkdir -p {params.filter_root:q}/Filtered_bams
    xenofilx run \
      --graft {input.graft_bams:q} \
      --host {input.host_bams:q} \
      --output {params.filter_root:q}/Filtered_bams \
      --mm-threshold {params.mm_threshold} \
      --unmapped-penalty {params.unmapped_penalty} \
      --threads {threads} \
      --recalculate-nm \
      --graft-ref {params.graft_ref:q} \
      --host-ref {params.host_ref:q} {"--bisulfite" if params.mode != "RNASEQ" else ""}
    for filtered_bam in {output.filtered_bams:q}; do
      test ! -L "$filtered_bam"
      test -s "$filtered_bam"
      enva run otter-core -- samtools quickcheck -v "$filtered_bam"
      enva run otter-core -- samtools index -@ {threads} "$filtered_bam"
    done
    for filtered_bai in {output.filtered_bais:q}; do
      test ! -L "$filtered_bai"
      test -s "$filtered_bai"
    done
    """
