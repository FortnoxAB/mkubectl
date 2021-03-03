# mkubectl
Run kubectl on multiple clusters using regexp for context name. 

## Usage

```
$ mkubectl --help
NAME:
   mkubectl - run kubectl command in multiple contexts

USAGE:
   mkubectl [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --context value, -c value  regexp kubectl context name
   --help, -h                 show help (default: false)
```


## Example

```
mkubectl -c site.*-dev get po

```
