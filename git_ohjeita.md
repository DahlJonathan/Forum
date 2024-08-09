

# Checking the Working Tree

Check the Status of the Working Tree:
```
git status
```
View Changes (Diff) in the Working Directory:
```
git diff
```
View Changes Staged for the Next Commit:
```
git diff --staged
```

# Creating and Moving Through Branches

List All Branches:
```
git branch
```
Create a New Branch:
```
git branch <new-branch-name>
```
Switch to Another Branch:
```
git checkout <branch-name>
```
or with Git 2.23 and later:
```
git switch <branch-name>
```
Create and Switch to a New Branch:
```
git checkout -b <new-branch-name>
```
or with Git 2.23 and later:
```
git switch -c <new-branch-name>
```

# Pulling and Pushing Changes

Pull Changes from Remote Repository:
```
git pull
```
Push Changes to Remote Repository:
```
git push
```
Push a Specific Branch to the Remote Repository:
```
git push origin <branch-name>
```

# Dealing with Conflicts

#### After Pulling Changes that Cause Conflicts:
#### Git will indicate which files have conflicts. You need to open these files and manually resolve the conflicts.

#### Mark a Conflict as Resolved:

After resolving the conflicts in a file, mark it as resolved:
```
git add <file-name>
```
Continue the Merge Process After Resolving Conflicts:
```
git commit
```
If you were in the middle of a rebase, continue with:
```
git rebase --continue
```
Abort a Merge if You Want to Start Over:
```
git merge --abort
```
Abort a Rebase if You Want to Start Over:
```
git rebase --abort
```

# Additional Useful Commands

View Commit History:
```
git log
```
View a More Compact Commit History:
```
git log --oneline
```
# Stash Changes:

Save changes to a stash:
```
git stash
```
Apply stashed changes:
```
git stash apply
```
Delete a Branch:
```
git branch -d <branch-name>
```
Force delete a branch:
```
git branch -D <branch-name>
```

# Working in Different Branches

Create and Switch to a New Branch:
```
git checkout -b <new-branch-name>
```
or with Git 2.23 and later:
```
git switch -c <new-branch-name>
```
Switch to an Existing Branch:
```
git checkout <branch-name>
```
or with Git 2.23 and later:
```
git switch <branch-name>
```
List All Branches:
```
git branch
```
# Pushing to the Main Branch

Switch to the Main Branch (usually main or master):
```
git checkout main
```
or:
```
git checkout master
```
Ensure the Main Branch is Up-to-Date:
```
git pull origin main
```
or:
```
git pull origin master
```
Merge Changes from Another Branch into Main:
```
git merge <branch-name>
```
Push Changes to the Remote Main Branch:
```
git push origin main
```
or:
```
git push origin master
```

# Changing to a Different Commit

View Commit History to Find the Commit Hash:
```
git log
```
Checkout a Specific Commit:
```
git checkout <commit-hash>
```
#### Note: This puts your repository in a "detached HEAD" state. You can make changes,
#### but it's often best to create a new branch if you intend to work from this commit:
```
git checkout -b <new-branch-name> <commit-hash>
```
Revert to a Specific Commit (Undoing All Commits After):
```
git reset --hard <commit-hash>
```
or to keep changes in the working directory:
```
git reset --soft <commit-hash>
```
Revert a Single Commit by Creating a New Commit:
```
git revert <commit-hash>
```

# Additional Tips

#### Rebasing Branches:

To apply your branch changes on top of another branch:
```
git rebase <base-branch>
```
For example, to rebase your feature branch onto the latest main:
```
git checkout <feature-branch>
git rebase main
```
#### Cherry-Picking a Commit:

To apply a specific commit from another branch:
```
git cherry-pick <commit-hash>
```
#### Squashing Commits:

To combine multiple commits into one:
```
git rebase -i <base-commit>
```
Then follow the instructions in the interactive rebase interface.


# Fetching

Fetch Updates from Remote Without Merging:
```
git fetch
```
Fetch from a Specific Branch:
```
git fetch origin <branch-name>
```

# Pushing from Another Branch

If you want to push changes from one branch to another branch in the remote repository, you can specify both the source and the destination branches.

#### Push Changes from a Local Branch to a Different Remote Branch:

Suppose you are on branch1 and you want to push its changes to the main branch on the remote repository.

    git push origin branch1:main

#### Force Push (if necessary, but use with caution):

If you need to force push (e.g., to overwrite the remote branch), you can add the --force or -f flag.
Be very careful with this, as it can overwrite changes in the remote branch.

    git push --force origin branch1:main

## Example Workflow

#### Here is an example of how you might pull changes from main into branch1 and then push changes from branch1 to main:

Checkout and Update branch1:

    git checkout branch1
    git pull origin main

#### This updates branch1 with the latest changes from main.

#### Resolve any conflicts if necessary:

Git will indicate if there are any merge conflicts that need to be resolved. After resolving conflicts, stage the resolved files:

    git add <resolved-file>
    git commit

Push Changes from branch1 to main:

    git push origin branch1:main
