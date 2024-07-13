#!/bin/bash
set -e

if [ ! -z $AZURE_TOKEN ] ; then
    # аутентификация в Azure
    BASIC_AUTH=$(printf "pat:%s""$REPO_PAT" | base64) 
    git config --global http."https://tfs.astralnalog.ru/tfs/Deimos/AstralEdo/_git/Astral.Edo.Backend.Go".extraHeader "Authorization: Basic $BASIC_AUTH"
fi

if [ ! -z $GITLAB_AUTH ] ; then 
    # аутентификация в GitLab
    git config --global url."https://$GITLAB_AUTH@git.astralnalog.ru".insteadOf "https://git.astralnalog.ru"
fi

# чтобы git не ругался на странную директорию
git config --global --add safe.directory /mnt

exec $@
