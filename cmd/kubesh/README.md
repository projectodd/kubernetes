# kubesh 

An interactive shell around `kubectl` providing tab-completion,
command history, readline key bindings, and an ability to "pin" a
resource type or name to which subsequent commands will apply. Common
options are "sticky" so that they need not be repeatedly specified
during a single session.

### Installation

Build `kubesh` from the root of this repository, as you would any
other Kubernetes command:
    
    cd ../..
    ./build-tools/run.sh make kubesh

The `kubesh` binary should then reside in a platform-specific
directory beneath `../../_output/dockerized/bin`

### Sample Session

![](http://i.imgur.com/Lg1zAnw.gif)

### Readline Shortcuts

* Normal mode

| Shortcut           | Comment                           |
| ------------------ | --------------------------------- |
| `Ctrl`+`a`         | Beginning of line                 |
| `Ctrl`+`b`         | Backward one character            |
| `Alt`+`b`          | Backward one word                 |
| `Ctrl`+`c`         | Send io.EOF                       |
| `Ctrl`+`d`         | Delete one character              |
| `Alt`+`d`          | Delete one word                   |
| `Ctrl`+`e`         | End of line                       |
| `Ctrl`+`f`         | Forward one character             |
| `Alt`+`f`          | Forward one word                  |
| `Ctrl`+`g`         | Cancel                            |
| `Ctrl`+`h`         | Delete previous character         |
| `Ctrl`+`i` / `TAB` | Command line completion           |
| `Ctrl`+`j`         | Line feed                         |
| `Ctrl`+`k`         | Cut text to the end of line       |
| `Ctrl`+`l`         | Clear screen                      |
| `Ctrl`+`m`         | Same as Enter key                 |
| `Ctrl`+`n`         | Next line (in history)            |
| `Ctrl`+`p`         | Prev line (in history)            |
| `Ctrl`+`r`         | Search backwards in history       |
| `Ctrl`+`s`         | Search forwards in history        |
| `Ctrl`+`t`         | Transpose characters              |
| `Alt`+`t`          | Transpose words (TODO)            |
| `Ctrl`+`u`         | Cut text to the beginning of line |
| `Ctrl`+`w`         | Cut previous word                 |
| `Backspace`        | Delete previous character         |
| `Alt`+`Backspace`  | Cut previous word                 |
| `Enter`            | Line feed                         |

* Search mode (via `Ctrl`+`s` or `Ctrl`+`r`)

| Shortcut                | Comment                                 |
| ----------------------- | --------------------------------------- |
| `Ctrl`+`s`              | Search forwards in history              |
| `Ctrl`+`r`              | Search backwards in history             |
| `Ctrl`+`c` / `Ctrl`+`g` | Exit Search Mode and revert the history |
| `Backspace`             | Delete previous character               |
| Other                   | Exit Search Mode                        |

* Complete Select mode (via double `TAB`)

| Shortcut                | Comment                                  |
| ----------------------- | ---------------------------------------- |
| `Ctrl`+`f`              | Move Forward                             |
| `Ctrl`+`b`              | Move Backward                            |
| `Ctrl`+`n`              | Move to next line                        |
| `Ctrl`+`p`              | Move to previous line                    |
| `Ctrl`+`a`              | Move to the first candicate in current line |
| `Ctrl`+`e`              | Move to the last candicate in current line |
| `TAB` / `Enter`         | Use the word on cursor to complete       |
| `Ctrl`+`c` / `Ctrl`+`g` | Exit Complete Select Mode                |
| Other                   | Exit Complete Select Mode                |
