#!/bin/bash

examples="rjoint_ex06_pullout seep_ex01_freesurf spo751_pressurised_cylinder spo754_strip_footing_collapse"
for ex in $examples; do
    echo
    echo
    echo "[1;32m>>> running $ex <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<[0m"
    cd $ex
    ./all.bash
    cd ..
done