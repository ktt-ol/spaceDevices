#!/usr/bin/python3

import csv

with open('oui.csv', newline='') as csvfile:
    oui = csv.reader(csvfile, delimiter=',', quotechar='"')
    skipFirst = True
    for row in oui:
        if skipFirst:
            skipFirst = False
            continue
        print('%s,%s' % (row[1], row[2]))
