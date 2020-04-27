import sys
from z3 import *
from argparse import ArgumentParser

parser = ArgumentParser()
parser.add_argument("-f", "--file",
    action="store_true", dest="z3file", default=False, 
    help="Z3 file to check")

#args = parser.parse_args()

s = Solver()
#s.from_file("/app/scripts/smt.z3")
s.from_file(sys.argv[1])
print(s.check())