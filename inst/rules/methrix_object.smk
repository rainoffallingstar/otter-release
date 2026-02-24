rule create_methrix_object :
  message:"Build beta matrix ..."
  input:
    expand(os.path.join(config["directories"]["methylation_call"], "{sample}_nsort.bismark.cov.gz"),sample = config["metadata"]["sample_ids"])
  output:
    os.path.join(config["directories"]["methylation_call"], "methrixh5","assays.h5"),
    os.path.join(config["directories"]["methylation_call"], "methrixh5","CpG_coverage.xlsx")

  params:
    mcall_dir = config["directories"]["methylation_call"],
    methrix_dir = os.path.join(config["directories"]["methylation_call"], "methrixh5"),
    genome = config["reference"]["genome_fasta"][config["workflow"]["species"]["name"].index(config["workflow"]["species"]["graft"])]
  threads:10
  shell:
    """
      methrix-cli process \
        --input {params.mcall_dir} \
        --output {params.methrix_dir} \
        --genome {params.genome} \
        --threads {threads}

    """
   




    
