# Opsilon - How to use

Opsilon comes with some basic terms to define, connect to and run container-native workflows defined as YAML files. These are called `workflows`. Workflow files end in `.ops.yaml`. Below is a full example of all the capabilities of a single workflow.

```yaml
# ID of job
id: example-full

# Help Description
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
    optional: true
  - name: arg3
    optional: true # If skipped in the CLI input phase, will default to an empty string [($arg3 == "") == true]

mount: true # If true, will mount a copy of the working directory as the volume of the stages

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
        ls -l
        echo "exportedArg=i_am_an_output" >> $OUTPUT
        ls -l $OUTPUT
        cat $OUTPUT
        mkdir testdir1
        echo $arg3 >> testdir1/test.txt
        echo "Stage Ended"
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
        echo "I am another stage"
        echo $exportedArg >> testdir2/test.txt
        ls -l
    artifacts:
      - testdir2 # Copies files inside it.
      - testdir2/test.txt # It has no effect to copy file twice.
  - stage: write a file
    id: writefile3
    needs: writefile # Will get the outputs of the stage with this ID. Comma Separated list of stage IDs
    clean: false # Enabling this will make this stage will not share a filesystem with the other stages. It will start with a clean /app as working directory.
    if: $exportedArg == "i_am_an_output" # Run only if the output of the step it needs is equal this string.
    script:
      - sh
      - -c
      - |
        mkdir testdir3
        echo "I am another stage"
        echo $exportedArg >> testdir3/test.txt
        ls -l
    artifacts:
      - testdir3 # Copies files inside it.
  - stage: read the file
    id: readfile
    needs: writefile3 # Comma Separated list of stage IDs
    if: $exportedArg == "wrong_output"
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
-  For private repositories, use `https://myuser:github_token@github.com/myprivateorg/myprivaterepo.git`


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

For now, let's include only the GIT repository in our config:
```sh
$> opsilon repo add --git -n examples -d examples -s examples/workflows -p https://github.com/jatalocks/opsilon.git -b main
$> opsilon repo list
Using config file: /Users/my.user/.opsilon.yaml
+------------------+--------------------------------+--------------------------------------+------+--------+--------------------+
|       NAME       |          DESCRIPTION           |               PATH/URL               | TYPE | BRANCH |     SUBFOLDER      |
+------------------+--------------------------------+--------------------------------------+------+--------+--------------------+
| example_repo_git | Contains example workflows     | https://github.com/jatalocks/opsilon | git  | main   | examples/workflows |
|                  | from a folder in a git         |                                      |      |        |                    |
|                  | repository                     |                                      |      |        |                    |
+------------------+--------------------------------+--------------------------------------+------+--------+--------------------+
```

And list available workflows:
```sh
$> opsilon list     
Using config file: /Users/my.user/.opsilon.yaml
Repository example_repo_git
Getting workflows from repo example_repo_git in location https://github.com/jatalocks/opsilon type git
+------------------+---------------+--------------------------------+-----------------------------+----------------+-------------+
|    REPOSITORY    |      ID       |          DESCRIPTION           |         IMAGES USED         |     INPUTS     | STAGE COUNT |
+------------------+---------------+--------------------------------+-----------------------------+----------------+-------------+
| example_repo_git | example-full  | this is an example workflow    | alpine:latest,ubuntu:latest | arg1,arg2,arg3 |           3 |
|                  |               | which includes all of opsilons |                             |                |             |
|                  |               | capabilities                   |                             |                |             |
| example_repo_git | example-small | this is an example workflow    | alpine:latest,ubuntu:latest | myinput        |           1 |
|                  |               | which contains some of         |                             |                |             |
|                  |               | opsilons capabilities          |                             |                |             |
+------------------+---------------+--------------------------------+-----------------------------+----------------+-------------+
```

Let's run. First let's look at the run command:
```sh
$> opsilon run --help
Run an available workflow

Usage:
  opsilon run [flags]

Flags:
  -a, --args stringToString   Comma separated list of key=value arguments for the workflow input (default [])
      --confirm               Start running without confirmation
  -h, --help                  help for run
      --kubernetes            Run in Kubernetes instead of Docker. You must be connected to a Kubernetes Context
  -r, --repo string           Repository Name
  -w, --workflow string       ID of the workflow to run

Global Flags:
      --config string   config file (default is $HOME/.opsilon.yaml)
```

As you can see, we can call a workflow from a certain repo, give it inputs and skip confirmation. But for now, let's run it without any argument and let it ask us what we need:

```sh
$> opsilon run       
Using config file: /Users/my.user/.opsilon.yaml
Use the arrow keys to navigate: ↓ ↑ → ← 

? Select Repo: 
  ▸ example_repo_git

✔ example_repo_git

Repository example_repo_git
Getting workflows from repo example_repo_git in location https://github.com/jatalocks/opsilon type git

Use the arrow keys to navigate: ↓ ↑ → ← 

Select Workflow
  ▶️ example-full (this is an example workflow which includes all of opsilons capabilities)
    example-small (this is an example workflow which contains some of opsilons capabilities)

▶️ example-full
You Chose: example-full
arg1 (): something
something
arg2 (defaultvalue): 
defaultvalue
arg3 (): another_thing
another_thing
--------- Running "example-full" with: ----------

arg1: something

arg2: defaultvalue

arg3: another_thing

? Run example-full? [Y/n] █

Run example-full: Y
Running in Parallel: writefile
[write a file:writefile] Evaluating If Statement: $arg3 != "", with the following variables: [{filename testValue} {onlyhere something} {arg1 something} {arg2 defaultvalue} {arg3 another_thing}]
[write a file:writefile] Running Stage with the following variables: [filename=testValue onlyhere=something arg1=something arg2=defaultvalue arg3=another_thing]
[write a file:writefile] Starting Stage
[write a file:writefile] Copying testdir1 To /Users/my.user/opsilon/testdir1
[write a file:writefile] Copied testdir1 To /Users/my.user/opsilon/testdir1
[write a file:writefile] stat /var/folders/cr/mbr1038j5s50tm9gvkq04vx00000gp/T/temp3900982779/test.txt: no such file or directory
Running in Parallel: writefile2
[write a file:writefile2] Evaluating If Statement: $exportedArg == "wrong_output", with the following variables: [{filename testValue} {arg1 something} {arg2 defaultvalue} {arg3 another_thing} {exportedArg i_am_an_output}]
[write a file:writefile2] Stage Skipped due to IF condition
Running in Parallel: readfile
[read the file:readfile] Evaluating If Statement: $exportedArg == "i_am_an_output", with the following variables: [{filename testValue} {arg1 something} {arg2 defaultvalue} {arg3 another_thing} {exportedArg i_am_an_output}]
[read the file:readfile] Stage Skipped due to needed stage skipped
```

And so we have chosen and ran our workflow! We could have done the same thing automatically without prompt with the following command:

```sh
$> opsilon run -r examples -w example-full --confirm -a "arg1=something,arg3=something" #arg2 has a default, we can choose to override.
```
