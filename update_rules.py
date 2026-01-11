#!/usr/bin/env python3
"""
Script to update Snakemake rules to use dot notation for nested config access
"""

import glob
import os
import re

# Mapping from flat config keys to nested config keys
CONFIG_MAPPING = {
    # Basic workflow fields
    "Mode": "workflow.mode",
    "mode": "workflow.mode",
    "userid": "workflow.userid",
    "jobid": "workflow.jobid",
    "species": "workflow.species.name",
    "graft": "workflow.species.graft",
    "host": "workflow.species.host",
    # Input fields
    "suffix": "input.suffix",
    "suffix2": "input.suffix2",
    "fastq_dir": "input.fastq_dir",
    "pdata_file": "input.pdata_file",
    # Adapters
    "error": "workflow.adapters.error",
    "trimSeq1": "workflow.adapters.seq1",
    "trimSeq2": "workflow.adapters.seq2",
    # Alignment parameters
    "C1": "workflow.alignment.c1",
    "C2": "workflow.alignment.c2",
    "T1": "workflow.alignment.t1",
    "T2": "workflow.alignment.t2",
    # Directory paths
    "workDir": "directories.work",
    "workflowDir": "output.workflow_dir",
    "analysisDir": "output.analysis_dir",
    "selfconfig": "directories.config",
    "qcDir": "directories.qc.main",
    "qcDir_before": "directories.qc.before",
    "qcDir_after": "directories.qc.after",
    "SID_log": "directories.sid_log",
    "trimDir": "output.trim_dir",
    "bsmapDir": "directories.bsmap.main",
    "bsmapDir_bamtmp": "directories.bsmap.temp",
    "outDir_mCall": "directories.methylation_call",
    "ourDirUmx": "directories.umx",
    "outdir_qualimap": "directories.qualimap",
    "outDir_mhap": "directories.mhap",
    "RData_folder": "directories.rdata",
    "DMR_folder": "directories.dmr",
    "rawDir": "output.raw_dir",
    "outDir_betaM": "directories.beta_matrix",
    "qc_summary": "directories.qc_summary",
    "logsummary": "directories.log_summary",
    "uxm_summary": "directories.uxm_summary",
    # Reference files
    "genomeFile": "reference.indices.genome",
    "genome_fasta": "reference.files.fasta",
    "genomeAnno": "reference.annotations.names",
    "cgGR_gz": "reference.files.cpg_sites",
    "CGI": "reference.files.cgi",
    # Sample and metadata
    "SIDs": "metadata.sample_ids",
    "user_email": "metadata.user_email",
    "group_levels": "metadata.group_levels",
    "pdx_pipeline": "metadata.pdx_pipeline",
    # Processing parameters
    "read1_5": "workflow.trim.read1_5",
    "read1_3": "workflow.trim.read1_3",
    "read2_5": "workflow.trim.read2_5",
    "read2_3": "workflow.trim.read2_3",
    "seq_deth": "workflow.trim.seq_deth",
    "fixed": "workflow.trim.fixed",
    # ClubCpG directories
    "clubcpg": "directories.clubcpg.main",
    "clubcpg_coverage_before": "directories.clubcpg.coverage",
    "clubcpg_model": "directories.clubcpg.model",
    "clubcpg_coverage_impute": "directories.clubcpg.impute",
    # Reference data
    "chrs": "reference.rnaseq.chromosomes",
    "rnaseq_gtf": "reference.rnaseq.gtf",
    "rnaseq_ref": "reference.rnaseq.ref",
    # Output files
    "methrixh5": "directories.methrix_h5",
    "GCbias": "directories.gc_bias",
}


def update_config_access(content):
    """Update config access patterns in the content"""
    # Pattern to match config["key"] or config['key']
    pattern = r'config\s*\[\s*["\']([^"\']+)["\']\s*\]'

    def replace_config(match):
        key = match.group(1)
        if key in CONFIG_MAPPING:
            nested_key = CONFIG_MAPPING[key]
            return f'config["{nested_key}"]'
        # Keep original if not in mapping
        return match.group(0)

    updated_content = re.sub(pattern, replace_config, content)
    return updated_content


def update_snakemake_rule(filepath):
    """Update a single Snakemake rule file"""
    print(f"Updating {filepath}...")

    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()

    # Check if file has any config access
    if "config[" not in content:
        print(f"  No config access found, skipping")
        return False

    # Update config access
    updated_content = update_config_access(content)

    # Check if anything changed
    if updated_content == content:
        print(f"  No changes needed")
        return False

    # Write updated content back
    with open(filepath, "w", encoding="utf-8") as f:
        f.write(updated_content)

    print(f"  Updated successfully")
    return True


def main():
    """Main function to update all Snakemake rule files"""
    print("Updating Snakemake rules to use nested config access...")
    print()

    # Find all .smk files
    rule_dirs = ["inst/root_rules", "inst/rootless_rules"]

    updated_count = 0
    total_count = 0

    for rule_dir in rule_dirs:
        if not os.path.exists(rule_dir):
            print(f"Directory {rule_dir} not found, skipping")
            continue

        print(f"Processing {rule_dir}...")
        smk_files = glob.glob(os.path.join(rule_dir, "*.smk"))
        total_count += len(smk_files)

        for smk_file in smk_files:
            if update_snakemake_rule(smk_file):
                updated_count += 1

        print()

    print(f"Summary:")
    print(f"  Total files: {total_count}")
    print(f"  Updated files: {updated_count}")
    print()
    print("Done!")


if __name__ == "__main__":
    main()
