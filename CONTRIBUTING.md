# Contribution Guide

All contributions are welcome _if_ they are in the scope of the project. If you're not sure about something, feel free to ask.

## Guidelines

- Make sure you are familiar with Go and React.
- Your contributions must be small and focused. If you want to add a new feature that requires substantial changes or additions to the codebase, please contact the dev first.
- Make sure your changes are in line with the project's goals (Create a feature request if you're unsure).
- Make sure your changes are well tested and do not introduce any new issues or regressions.
- You should try and make your changes against the **most active branch**, which is usually the `main` branch but
  may be different when a new version is being developed.

## How to contribute

1. Create an issue before starting work on a feature or a bug fix.
2. Fork the repository, clone it, and create a new branch.

	```shell
	# Clone your fork of the repo
	git clone https://github.com/<your-username>/seanime.git
	# Navigate to the directory
	cd seanime
	# Assign to a remote called "upstream"
	git remote add upstream https://github.com/5rahim/seanime.git
	```

3. Get the latest changes from the original repository.

	```shell
	git fetch --all
	git rebase upstream/main
	```

4. Create a new branch for your feature or bug fix off of the `main` branch.

	```shell
	git checkout -b <feature-branch> main
	```

5. Make your changes, test and commit them.

6. Locally rebase your changes on top of the latest changes from the original repository.

	```shell
	git pull --rebase upstream main
	```

7. Push your changes to your fork.

	```shell
	git push -u origin <feature-branch>
	```

8. Create a pull request to the `main` branch of the original repository.

9. Wait for the maintainers to review your pull request.

10. Make changes if requested.

11. Once your pull request is approved, it will be merged.

12. Keep your fork in sync with the original repository.

	```shell
	git fetch --all
	git checkout main
	git rebase upstream/main
	git push -u origin main
	```

## Areas

[Issues](https://github.com/5rahim/seanime/issues?q=is%3Aissue+is%3Aopen+label%3A%22open+to+contribution%22)
