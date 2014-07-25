# Duh

A `du`-like tool that gathers file and directory size statistics and plots to a histogram in realtime.

Install with `go install github.com/mtibben/duh`

Run with `duh`

## Example
```
$ duh ~
                 ####     6.3G   Dropbox
                 ####     6.3G   Projects
                 ####     6.5G   Pictures
                #####     9.2G   Downloads
        #############    23.8G   Library
    #################    30.8G   Music
 ####################    38.2G   VirtualBox VMs
               TOTAL:   137.7G   (827784 files)
```

## TODO:
 - interactive expandable tree
 - colour
 - globbing
