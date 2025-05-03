#!/bin/zsh
printf "Shit broke! Killing SDL2."
sshpass -p 'tina' ssh root@192.168.1.16 "kill  \$(pidof dlv)" > /dev/null 2>&1
sshpass -p 'tina' ssh root@192.168.1.16 "kill  \$(pidof nextui-sdl2)" > /dev/null 2>&1
