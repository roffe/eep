## eep completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(eep completion bash)

To load completions for every new session, execute once:

#### Linux:

	eep completion bash > /etc/bash_completion.d/eep

#### macOS:

	eep completion bash > /usr/local/etc/bash_completion.d/eep

You will need to start a new shell for this setup to take effect.


```
eep completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -c, --chip uint8     chip type (default 66)
  -e, --erase          erase before write (default false
  -o, --org uint8      chip org (default 8)
  -p, --port string    com port (default "COM8")
  -s, --size uint16    chip size (default 512)
  -x, --xor bytesHex   xor output (default 00)
```

### SEE ALSO

* [eep completion](eep_completion.md)	 - Generate the autocompletion script for the specified shell

