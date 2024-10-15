### Get version 

This action allows you to get version of the software from different sources:
1. Trigger a Jenkins job
2. Wait for a Jenkins job to finish
3. Get the status of a Jenkins job
4. Get the console output of a Jenkins job

## Example

```yaml
jobs:
  run-jenkins-job:
    runs-on: ubuntu-latest
    steps:
      - name: Get Latest Ubuntu version
        id: get-ubuntu-version
        uses: scylladb-actions/get-version@v0.1.0
        with:
          source: dockerhub-imagetag
          repo: ubuntu
          filters: LAST.LAST.LAST

      - name: Print the version
        run: echo "The version is ${{ fromJson(steps.get-ubuntu-version.outputs.versions)[0] }}"
```

### Cli usage

CLI Arguments:
* -filters string. Filters to apply to versions. Example: "LAST.*.*"
* -mvn-artifact-id string. Artifact ID to search on the maven
* -mvn-group string. Artifact group to search on the maven
* -out-no-prefix. Remove prefix from output
* -out-reverse-order. Reverse order
* -out-format string. Output type: json, yaml, text (default "text")
* -prefix string. Version prefix
* -repo string. Repository name. Examples for dockerhub: ubuntu or alpine/git; for github: golang/go or scylladb/scylla
* -source string. Version source, one of: dockerhub-imagetag, maven-artifact, github-release, github-tag

```bash
get-versions --source dockerhub-imagetag --repo ubuntu
```