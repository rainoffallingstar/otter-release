def rnaseq_splicing_bams(wildcards):
    pdx_mode = str(config["metadata"]["pdx_pipeline"]).strip().lower() in {"1", "true", "yes"}
    if pdx_mode:
        return expand(
            os.path.join(config["directories"]["bsmap"]["main"], "Filtered_bams", "{sample}_" + config["workflow"]["species"]["graft"] + "_Filtered.bam"),
            sample=config["metadata"]["sample_ids"],
        )
    return expand(
        os.path.join(config["directories"]["bsmap"]["main"], "{sample}_{species}.bam"),
        sample=config["metadata"]["sample_ids"],
        species=config["workflow"]["species"]["name"],
    )

rule rnaseq_splicing:
    message: "RNA Splicing ..."
    input:
        bams=rnaseq_splicing_bams,
        pdata=os.path.join(config["directories"]["selfconfig"], "pdata.xlsx"),
        gtf=config["reference"]["rnaseq"]["gtf"][config["workflow"]["species"]["name"].index(config["workflow"]["species"]["graft"])]
    output:
        marker=os.path.join(config["directories"]["bsmap"]["main"], "RNASplicing", "RNASplicing_success.txt")
    params:
        run_dir=config["directories"]["bsmap"]["main"],
        seqlengthQC=config["directories"]["qc"]["main"],
        log_dir=os.path.join(config["directories"]["bsmap"]["main"], "RNASplicing"),
        pdxmode=(1 if str(config["metadata"]["pdx_pipeline"]).strip().lower() in {"1", "true", "yes"} else 0)
    threads: 20
    run:
        if config["metadata"]["group_levels"] >= 2:
            shell(
                """
                enva run xdxtools-core -- \
                  gomats run \
                  --root {params.run_dir:q} \
                  --threads {threads} \
                  --pdata {input.pdata:q} \
                  --seqlengthQC {params.seqlengthQC:q} \
                  --gtf {input.gtf:q} \
                  --pdxmode {params.pdxmode}
                printf '%s\n' 'RNASplicing_DONE' > {output.marker:q}
                """
            )
        else:
            shell(
                """
                mkdir -p {params.log_dir:q}
                printf '%s\n' 'RNASplicing_NOTRUN' > {output.marker:q}
                """
            )
