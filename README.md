# Summary

This tool calculates some basic statistics from Half-Life 1 ("valve game") log
file.

# Usage Example

```
$ ./vlp -h
Usage of ./vlp:
  -logfile string
        path to logfile (default "L0226001.log")
```

```
$ ./vlp ./L0226001.log 
INFO[0000] map played: stalkyard                        
INFO[0000] Pluto summary:                              
INFO[0000]       playtime: 14135s                           
INFO[0000]       frags: 873                                 
INFO[0000]       suicides: 51                               
INFO[0000] Moon summary:                                
INFO[0000]       playtime: 11990s                           
INFO[0000]       frags: 563                                 
INFO[0000]       suicides: 39                               
INFO[0000] Mercury summary:                           
INFO[0000]       playtime: 11306s                           
INFO[0000]       frags: 343                                 
INFO[0000]       suicides: 16                               
INFO[0000] Uranus summary:                            
INFO[0000]       playtime: 11883s                           
INFO[0000]       frags: 521                                 
INFO[0000]       suicides: 13                               
INFO[0000] Saturn summary:                               
INFO[0000]       playtime: 11249s                           
INFO[0000]       frags: 445                                 
INFO[0000]       suicides: 29                               
INFO[0000] who kills whom:                              
who,Pluto,Moon,Mercury,Uranus,Saturn
Pluto,51,257,213,233,170
Moon,165,39,136,147,115
Mercury,91,94,16,92,66
Uranus,126,157,133,13,105
Saturn,107,121,95,122,29
INFO[0000] who kills with what:
what,Pluto,Moon,Mercury,Uranus,Saturn
357,23,9,1,2,16
satchel,18,38,6,5,30
tank,13,3,25,0,3
crowbar,6,79,68,53,22
rpg_rocket,76,49,4,13,12
9mmAR,177,90,42,86,78
tau_cannon,15,0,1,0,8
9mmhandgun,76,13,73,40,25
shotgun,257,197,102,255,139
snark,3,2,2,0,0
tripmine,95,23,16,7,17
crossbow,60,12,0,38,17
grenade,54,48,3,22,78
INFO[0000] who is killed of what:
what,Pluto,Moon,Mercury,Uranus,Saturn
357,7,14,9,13,8
satchel,23,17,20,27,10
tank,11,6,8,10,9
crowbar,38,44,37,60,49
rpg_rocket,33,30,37,29,25
9mmAR,83,107,118,95,70
tau_cannon,1,9,6,6,2
9mmhandgun,45,62,31,47,42
shotgun,166,216,203,194,171
snark,0,1,2,2,2
tripmine,15,46,27,46,24
crossbow,20,40,33,18,16
grenade,47,37,46,47,28
```
