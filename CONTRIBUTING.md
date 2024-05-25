# Contribution Guide

All contributions are welcome. If you're not sure about something, feel free to ask.

## Guidelines

Note that you should try and make your changes against the **most active branch**, which is usually the `main` branch but
may be different when a new minor version is being developed.

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

Here are some areas where you can contribute:

### Improve existing features

You can help improve the existing features by fixing bugs, adding new options, or general code improvements.

### Clients

You can help broaden the support for different platforms by helping create new clients for them, be it a port of the existing web client or a new one built from scratch.

### Code quality

You can help improve the code quality by refactoring some parts of the codebase using better standards and submitting a pull request.

### Documentation

Although the documentation has many details about the specific features, it may lack some important information
about specific use cases and platform-specific details. You can help by adding more details to the documentation.
- Installation on different platforms.
- Configuration options.
- Remote access and security.

