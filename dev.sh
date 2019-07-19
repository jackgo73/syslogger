git config user.name "JackGo"
git config user.email "jackgo73@outlook.com"
git checkout develop
git add -A
git commit -m 'update'
git push origin develop
git checkout master
git merge develop
git push origin master
git checkout develop