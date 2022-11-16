# Opsilon - How to use

Opsilon comes with some basic terms to define, connect to and run container-native workflows defined as YAML files. These are called `workflows`. Workflow files end in `.ops.yaml `. Below is a full example of all the capabilities of a single workflow.

```yaml
# ID of job.
id: example-full

# Help Description.
description: this is an example workflow which includes all of opsilons capabilities

# Global Docker Image. Used if no Stage specific image is specified.
image: alpine:latest

# Global Environment Variables for use inside the containers.
env:
  - name: filename
    value: testValue

# Inputs for users to enter in the CLI when using 'run'
input:
  - name: arg1
  - name: arg2
    default: defaultvalue
  - name: arg3
    optional: true # If skipped in the CLI input phase, will default to an empty string [($arg3 == "") == true]

# Stages Rules
# 1. All stages will run in parallel unless they have a "needs" field
stages:
  - stage: write a file # Name of the stage. These can be non-unique.
    id: writefile # ID of the stage. Used for 'outputs' and 'needs'. These need to be unique.
    image: ubuntu:latest # Override global image for this stage only.
    env: # Stage specific environment variables
      - name: onlyhere
        value: something
    # 'If' statements support normal mathematical expressions. 
    # Variables can be any variable available to the stage (Using '$' sign).
    if: $arg3 != "" # Skip if arg3 is empty. Run if not. 
    script: # Array of arguments to the container. $OUTPUT contains an output file. every key=value here will be available for export.
      - sh
      - -c
      - |
        echo "Starting Stage"
        echo "exportedArg=i_am_an_output" >> $OUTPUT
        mkdir testdir1
        echo $arg3 >> testdir1/test.txt
    artifacts: # Will be saved to the Working Directory where opsilon CLI was run from.
      - testdir1
      - test.txt # Will not exist, the runner will ignore this and print a warning.
  - stage: write a file
    id: writefile2
    needs: writefile # Will get the outputs of the stage with this ID. Comma Separated list of stage IDs
    clean: true # Enabling this will make this stage will not share a filesystem with the other stages. It will start with a clean /app as working directory.
    if: $exportedArg == "wrong_output" # Run only if the output of the step it needs is equal this string.
    script:
      - sh
      - -c
      - |
        mkdir testdir2
        echo $exportedArg >> testdir2/test.txt
        ls -l
    artifacts:
      - testdir2 # Copies files inside it.
      - testdir2/test.txt # It has no effect to copy file twice.
  - stage: read the file
    id: readfile
    needs: writefile,writefile2 # Comma Separated list of stage IDs
    if: $exportedArg == "i_am_an_output"
    script:
      - sh
      - -c
      - cat testdir1/test.txt
```

Now that you have a functioning workflow, you can run it. But before that, you need to define it in a repository.

```
A "Repository" is a folder or a git repository containing .ops.yaml files.
```
You can define repositories for use in your local CLI by editing the .opsilon configuration file manually, or by using the CLI `repo` command.

The default configuration file can be found in: `$HOME/.opsilon.yaml`.
```yaml
repositories:
  - name: example_repo_folder
    description: Contains example workflows from a folder on the computer
    location:
      path: /myfolder/opsilon/examples/workflows
      type: folder
  - name: example_repo_git
    description: Contains example workflows from a folder in a git repository
    location:
      path: https://github.com/jatalocks/opsilon
      type: git
      subfolder: examples/workflows # Optional. Will take files that contain this in their Path. 
      branch: main # If omitted, will fetch default branch 
```

The `repo` command includes useful commands to edit this file. Any argument missing from the command will be prompted to you as an input:

```sh
Operate on workflow repositories

Usage:
  opsilon repo [command]

Available Commands:
  add         Add a workflow repo
  delete      Delete a repo from your local config
  list        List all available repositories

Flags:
  -h, --help   help for repo

Global Flags:
      --config string   config file (default is $HOME/.opsilon.yaml)

Use "opsilon repo [command] --help" for more information about a command.
```