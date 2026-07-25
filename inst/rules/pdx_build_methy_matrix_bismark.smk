rule bismark_methylation_extractor :
  message:"Build beta matrix ..."
  input:
    bam_sorted = lambda wildcards:os.path.join(config["directories"]["bsmap"]["main"],"Filtered_bams" ,f"{wildcards.sample}_{config['workflow']['species']['graft']}_Filtered.bam")
  output:
    os.path.join(config["directories"]["methylation_call"], "{sample}_nsort.bismark.cov.gz"),
    os.path.join(config["directories"]["bsmap"]["main"], "{sample}_nsort.bam")
  params:
    mem_size = "34G",
    mcall_dir = config["directories"]["methylation_call"],
    genomeFile = config["reference"]["indices"]["genome"][config["workflow"]["species"]["name"].index(config["workflow"]["species"]["graft"])],
    bam_nsorted = lambda wildcards:os.path.join(config["directories"]["bsmap"]["main"], f"{wildcards.sample}_nsort.bam")
  threads:5
  shell:
    """
    # Sort by read name and filter unpaired reads using pairbam
    enva run otter-core -- samtools sort  -@ {threads} -n -o {params.bam_nsorted}.tmp {input.bam_sorted}
    pairbam {params.bam_nsorted}.tmp {params.bam_nsorted}
    rm -f {params.bam_nsorted}.tmp

    enva run otter-core -- bismark_methylation_extractor --paired-end --gzip \
      --output_dir {params.mcall_dir} \
      --comprehensive --merge_non_CpG --bedGraph --multicore {threads} \
      --buffer_size {params.mem_size} \
      {params.bam_nsorted}

    """
   




    
