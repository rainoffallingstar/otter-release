rule prepare_methrix_reference_cpg:
  message:"Prepare methrix reference CpG (.ron) ..."
  input:
    genome_ref = lambda wildcards: config["reference"]["genome_fasta"][config["workflow"]["species"]["name"].index(config["workflow"]["species"]["graft"])]
  output:
    os.path.join(config["directories"]["methylation_call"], "methrixh5", "reference_cpgs.ron")
  params:
    genome_key = config["workflow"]["species"]["graft"]
  shell:
    """
      mkdir -p "$(dirname {output})"
      ref="{input.genome_ref}"
      out_dir="$(dirname {output})"
      key="{params.genome_key}"

      # Copy genome annotation GTF if available in reference directory.
      # Priority: <species>.gtf(.gz), then unique/first *.gtf(.gz).
      ref_dir="$(dirname "$ref")"
      gtf_src=""
      for c in "$ref_dir/${key}.gtf" "$ref_dir/${key}.gtf.gz"; do
        if [[ -f "$c" ]]; then
          gtf_src="$c"
          break
        fi
      done
      if [[ -z "$gtf_src" ]]; then
        gtf_any=("$ref_dir"/*.gtf "$ref_dir"/*.gtf.gz)
        for c in "${gtf_any[@]}"; do
          if [[ -f "$c" ]]; then
            gtf_src="$c"
            break
          fi
        done
      fi
      if [[ -n "$gtf_src" ]]; then
        cp -f "$gtf_src" "$out_dir/"
      fi

      if [[ "$ref" == *.ron || "$ref" == *.con ]]; then
        cp -f "$ref" "{output}"
      elif [[ -f "${ref}.ron" ]]; then
        cp -f "${ref}.ron" "{output}"
      else
        stem="$(basename "$ref")"
        stem="${stem%.gz}"
        stem="${stem%.fa}"
        stem="${stem%.fasta}"
        stem="${stem%.fna}"
        ref_dir="$(dirname "$ref")"
        near_ron="${ref_dir}/${stem}.ron"

        if [[ -f "$near_ron" ]]; then
          cp -f "$near_ron" "{output}"
        else
          methrix-cli extract-cpgs --genome "$ref" --output "{output}" || \
          methrix-cli extract-cp-gs --genome "$ref" --output "{output}"
        fi
      fi
    """

rule create_methrix_object :
  message:"Build beta matrix ..."
  input:
    expand(os.path.join(config["directories"]["methylation_call"], "{sample}_nsort.bismark.cov.gz"),sample = config["metadata"]["sample_ids"]),
    os.path.join(config["directories"]["methylation_call"], "methrixh5", "reference_cpgs.ron")
  output:
    os.path.join(config["directories"]["methylation_call"], "methrixh5","assays.h5"),
    os.path.join(config["directories"]["methylation_call"], "methrixh5","CpG_coverage.xlsx"),
    os.path.join(config["directories"]["methylation_call"], "methrixh5","CpG_annotation_report.xlsx")

  params:
    mcall_dir = config["directories"]["methylation_call"],
    methrix_dir = os.path.join(config["directories"]["methylation_call"], "methrixh5"),
    genome = os.path.join(config["directories"]["methylation_call"], "methrixh5", "reference_cpgs.ron")
  threads:10
  shell:
    """
      methrix-cli process \
        --input "{params.mcall_dir}" \
        --output "{params.methrix_dir}" \
        --genome "{params.genome}" \
        --annotation-dir "{params.methrix_dir}" \
        --threads {threads}

    """
   




    
