# test_init

This is an xdxtools project initialized with beaverflow structure.

## Workflow Mode: RRBS
## Engine Type: rootless

## Directory Structure

- config/          - Configuration files
- data/            - Input data (FASTQ files, pdata)
- envs/            - Conda environments
- inst/            - Genome reference files
- R/               - R/Python scripts
- rules/           - Snakemake rules
- saveRDS/         - R intermediate results
- temp/            - Temporary files
- userspace/       - User workspace
- workflows/       - Slurm/K8s job configs
- www/             - Web resources
- *.snakemake      - Snakemake workflow files

## Usage

1. Copy your FASTQ files to data/
2. Place your pdata file (Excel/CSV) in data/
3. Create config file: xdxtools config create --mode RRBS --output config/config.yaml
4. Run workflow: xdxtools run --config config/config.yaml

## Snakemake Workflows

Available workflows:
- BeaverBS_step1/2/3.snakemake   - RRBS/WGBS bisulfite sequencing
- BeaverPDX_step1/2/3.snakemake  - PDX bisulfite sequencing
- BeaverRNA_step1/2.snakemake    - RNA-seq analysis
- BeaverRNASEQPDX_step1/2/3.snakemake - PDX RNA-seq analysis

