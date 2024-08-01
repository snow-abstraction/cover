# cover

Using the SCP "Set Cover Problem" as a context to play with Go.

The initial plan is to mix some algorithms to build a solver for
weighted covering problems with strictly positive costs. I haven't
decided if the focus will be on general covers or exact covers (also known
as the "set partitioning problem").

# Dev Note

While this is a Go project, a Python program is used to generate test data.
This program independently solves SCP instances so we can verify that equally
good solutions are found by our code. For my Ubuntu system, here is a simple
way to get started:

```
cd tools
sudo apt install libffi-dev # install requirement needed by next line
pip install -i requirements.txt
cd ..
go run cmd/generate_test_instances_and_solutions/main.go -verbose
```

(A less hacky setup would be use a container or Python virtual environment.)

# Project Note

As of August 2023, I have focused little on Go and thus I am unsure of the
point of this project since then. The primary reason for this project was to
become better at Go for a project at work. But that project has been paused
and my focus returned a Java code base.

# License

This project is under [AGPL-3.0-only](LICENSE) license.
