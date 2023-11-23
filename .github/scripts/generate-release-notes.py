# Credit: https://github.com/metafates/mangal
import pathlib as pl

IN = "CHANGELOG.md"
OUT = "whats-new.md"

# get script path
script_path = pl.Path(__file__).resolve()

# get project root (3 levels up)
project_root = script_path.parent.parent.parent

# get changelog path
changelog_path = pl.Path(project_root, IN)

# get changelog content
with open(changelog_path, "r") as f:
    changelog = f.read()

    # we need to extract everything between the first and the second header with tags should not remove other header tags
    changelog = changelog.split("## ")[1]  # remove everything before the first header
    changelog = changelog.split("## ")[0]  # remove everything after the second header

    # remove the first line
    changelog = "\n".join(changelog.split("\n")[1:])

    # trim newlines
    changelog = changelog.strip()

    # write to file
    with open(pl.Path(project_root, OUT), "w") as ft:
        ft.write(changelog)