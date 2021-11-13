# py

[![License](https://img.shields.io/github/license/FollowTheProcess/py)](https://github.com/FollowTheProcess/py)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/py)](https://goreportcard.com/report/github.com/FollowTheProcess/py)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/py?logo=github&sort=semver)](https://github.com/FollowTheProcess/py)
[![CI](https://github.com/FollowTheProcess/py/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/py/actions?query=workflow%3ACI)

 (Approximate) port of Brett Cannon's [python-launcher] for Unix to Go.

* Free software: MIT License

## Project Description

This is primarily a learning exercise for now, you should use the original [python-launcher].

**Description of the original [python-launcher]:**

> Taken directly from the official [README]

Launch your Python interpreter the lazy/smart way!

This project is an implementation of the `py` command for Unix-based platforms
(with some potential experimentation for good measure 😉).

The goal is to have `py` become the cross-platform command that Python users
typically use to launch an interpreter while doing development. By having a
command that is version-agnostic when it comes to Python, it side-steps
the "what should the `python` command point to?" debate by clearly specifying
that upfront (i.e. the newest version of Python that can be found). This also
unifies the suggested command to document for launching Python on both Windows
as Unix as `py` has existed as the preferred
[command on Windows](https://docs.python.org/3/using/windows.html#launcher)
since 2012 with the release of Python 3.3.

Typical usage would be:

``` shell
py -m venv .venv
py ...  # Normal `python` usage.
```

This creates a virtual environment in a `.venv` directory using the latest
version of Python installed. Subsequent uses of `py` will then use that virtual
environment as long as it is in the current (or higher) directory; no
environment activation required (although the Python Launcher supports activated
environments as well)!

A non-goal of this project is to become the way to launch the Python
interpreter _all the time_. If you know the exact interpreter you want to launch
then you should launch it directly; same goes for when you have
requirements on the type of interpreter you want (e.g. 32-bit, framework build
on macOS, etc.). The Python Launcher should be viewed as a tool of convenience,
not necessity.

### Why are you reimplementing it in Go?

* I don't know, fun I guess?
* I love the original [python-launcher] and I love Go, so why not combine them!
* Learning and stuff

## Installation

## Quickstart

[python-launcher]: https://github.com/brettcannon/python-launcher
[README]: https://github.com/brettcannon/python-launcher/blob/main/README.md
