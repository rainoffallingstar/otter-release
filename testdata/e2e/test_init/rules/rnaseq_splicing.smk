rule rnaseq_splicing:
    message: "RNA Splicing ..."
    input:
        expand(os.path.join(config["directories"]["bsmap"]["main"], "{sample}_{species}.bam"), sample=config["SIDs"], species=config["species"])
    output:
        os.path.join(config["directories"]["bsmap"]["main"], "RNASplicing", "RNASplicing_success.txt")
    params:
        run_dir=config["directories"]["bsmap"]["main"],
        pdata=os.path.join(config["selfconfig"], "pdata.xlsx"),
        seqlengthQC=config["directories"]["qc"]["main"],
        gtf=config["rnaseq_gtf"][config["species"].index(config["workflow"]["species"]["graft"])],
        log_marker=os.path.join(config["directories"]["bsmap"]["main"], "RNASplicing", "RNASplicing_success.txt"),
        log_dir=os.path.join(config["directories"]["bsmap"]["main"], "RNASplicing"),
        pdxmode=(0 if not config["metadata"]["pdx_pipeline"] else 1)
    threads: 20
    run:
        if config["group_levels"] >= 2:
            shell(
                """
                enva run base \
                  Rscript R/RNA_Splicing.R \
                  --root {params.run_dir} \
                  --threads {threads} \
                  --pdata {params.pdata} \
                  --seqlengthQC {params.seqlengthQC} \
                  --gtf {params.gtf} \
                  --pdxmode {params.pdxmode}
                echo "RNASplicing_DONE" > {params.log_marker}
                """
            )
        else:
            shell(
                """
                mkdir -p {params.log_dir}
                echo "RNASplicing_NOTRUN" > {params.log_marker}
                """
            )
