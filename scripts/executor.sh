#!/usr/bin/env bash

#line="2022-08-20 09:54:10 -> 1"
#ff1=$(echo $line | awk -F '->' '{print $1}')
#ff2=$(echo $line | awk -F '->' '{print $2}')
#echo $ff1
#echo $ff2
#dates=()
#dates=( "${dates[@]}" "$ff1" )
#dates=( "${dates[@]}" "$ff1" )
#echo "${dates[@]}"
#echo ${#dates[@]}
#exit 1

output=$(./rewriting-history)
dates=()
commits=()
while IFS= read -r line; do
    filter=$(echo "$line" | grep -e '->')
    if [ -z "$filter" ]; then
      echo "$line"
    else
      d=$(echo $line | awk -F ' -> ' '{print $1}')
      c=$(echo $line | awk -F ' -> ' '{print $2}')
      dates=( "${dates[@]}" "$d" )
      commits=( "${commits[@]}" "$c" )
    fi
done <<< "$output"

echo "ready to commit"
for i in "${!dates[@]}"; do
  echo "executing: git commit --amend --date=\"${dates[$i]}\" -C ${commits[$i]} --no-edit"
#  git commit --amend --date="${dates[$i]}" -C ${commits[$i]} --no-edit
  git filter-branch --env-filter '
  if [ $GIT_COMMIT = "${commits[$i]}" ]; then
    export GIT_AUTHOR_DATE="${dates[$i]}"
    export GIT_COMMITTER_DATE="${dates[$i]}"
  fi
  '
done

#  git filter-branch --env-filter \
#      'if [ $GIT_COMMIT = efbc488b ]
#       then
#           export GIT_AUTHOR_DATE="Fri Jan 2 21:38:53 2009 -0800"
#           export GIT_COMMITTER_DATE="Sat May 19 01:01:01 2007 -0700"
#       fi'
