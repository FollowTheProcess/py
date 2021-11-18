---
title: PY
section: 1
header: Py
footer: Python Launcher (Experimental Go Port)
date: CURRENT_DATE
---

# NAME

py - launch a Python interpreter

# SYNOPSIS

**py** [**-[X]/[X.Y]**] ...

# DESCRIPTION

**py** launches the most appropriate Python interpreter it can find. It is meant
to act as a shorthand for launching **python** without having to think about
_which_ Python interpreter is the most desired. The Python Launcher is not meant
to substitute all ways of launching Python, e.g. if a specific Python
interpreter is desired then it is assumed it will be directly executed.

This **py** (https://github.com/FollowTheProcess/py) is an experimental port of the original,
written in Go.

Full credit to the original python-launcher by Brett Cannon available here: https://github.com/brettcannon/python-launcher

# SPECIFYING A PYTHON VERSION

If a command-line option is provided in the form of **-X** or **-X.Y** where _X_
and _Y_ are integers, then that version of Python will be launched
(if available). For instance, providing **-3** will launch the newest version of
Python 3 while **-3.6** will try to launch Python 3.6.

# SEARCHING FOR PYTHON INTERPRETERS

This is where this version differs slightly in behaviour from the original.

When no command-line arguments are provided to the launcher, what is deemed the
most "appropriate" interpreter is searched for as follows:

1. An activated virtual environment (launched immediately if available)
2. A **.venv** directory in the current working directory (launched immediately if available)
3. A **venv** directory in the current working directory (launched immediately if available)
4. If a file path is provided as the first argument, look for a shebang line
   containing **/usr/bin/python**, **/usr/local/bin/python**,
   **/usr/bin/env python** or **python** and any version specification in the
   executable name is treated as a version specifier (like with **-X**/**-X.Y**
   command-line options)
5. Check for any appropriate environment variable (see **ENVIRONMENT**)
6. Search **PATH** for all **pythonX.Y** executables
7. Launch the newest version of Python (while matching any version restrictions
   previously specified)

All unrecognized command-line arguments are passed on to the launched Python
interpreter.

# OPTIONS

**--help**
: Print a help message and exit; must be specified on its own.

**--list**
: List all known interpreters (except activated virtual environment);
must be specified on its own.

**-[X]**
: Launch the latest Python _X_ version (e.g. **-3** for the latest
Python 3). See **ENVIRONMENT** for details on the **PY_VERSION[X]** environment
variable.

**-[X.Y]**
: Launch the specified Python version (e.g. **-3.9** for Python 3.9).

# ENVIRONMENT

**PY_PYTHON**
: Specify the version of Python to search for when no Python
version is explicitly requested (must be formatted as 'X.Y'; e.g. **3.9** to use
Python 3.9 by default).

**PYLAUNCH_DEBUG**
: Log details to stderr about how the Launcher is operating.

**VIRTUAL_ENV**
: Path to a directory containing virtual environment to use when no
Python version is explicitly requested; typically set by
activating a virtual environment.

**PATH**
: Used to search for Python interpreters.

# AUTHORS

Original python-launcher: Copyright © 2018 Brett Cannon, Licensed under MIT.

This version authored by Tom Fleet, same license.

# HOMEPAGE

https://github.com/FollowTheProcess/python-launcher/

# SEE ALSO

python(1), python3(1), python-launcher(1)
