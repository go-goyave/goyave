#!/bin/bash

set -e

usage() { echo "Usage: $0 <project_name>" 1>&2; exit 1; }

if [ $# -ne 1 ]; then
    usage
fi

if [ -d $1 ]; then
  echo -e "\033[31m\033[1mError: \033[0mdirectory \"\033[37m$1\033[0m\" already exists."
  exit 1
fi

echo -e "\033[36m\033[1m                                                             
  ,ad8888ba,                                                                  
 d8\"'    \`\"8b                                                                 
d8'                                                                           
88              ,adPPYba,   8b       d8  ,adPPYYba,  8b       d8   ,adPPYba,  
88      88888  a8\"     \"8a  \`8b     d8'  \"\"     \`Y8  \`8b     d8'  a8P_____88  
Y8,        88  8b       d8   \`8b   d8'   ,adPPPPP88   \`8b   d8'   8PP\"\"\"\"\"\"\"  
 Y8a.    .a88  \"8a,   ,a8\"    \`8b,d8'    88,    ,88    \`8b,d8'    \"8b,   ,aa  
  \`\"Y88888P\"    \`\"YbbdP\"'       Y88'     \`\"8bbdP\"Y8      \"8\"       \`\"Ybbd8\"'  
                                d8'                                           
                               d8'                                            
\033[0m"
echo -e "\033[37m------------------------------------------------------------------------------\n"

echo -e "\033[92m\033[1mThank you for using Goyave!\033[0m"
echo -e "If you like the framework, please consider supporting me on Patreon:\n\033[37mhttps://www.patreon.com/bePatron?u=25997573\033[0m\n"

echo -e "\033[37m------------------------------------------------------------------------------\n"

echo -e "\033[1mDownloading template project...\033[0m"
curl -LOk https://github.com/System-Glitch/goyave-template/archive/master.zip

echo -e "\033[1mUnzipping...\033[0m"
unzip -q master.zip
rm master.zip
echo -e "\033[1mSetup...\033[0m"
mv goyave-template-master $1
cd $1
find ./ -type f \( -iname \*.go -o -iname \*.mod -o -iname \*.json \) -exec sed -i.bak "s/goyave_template/$1/g" {} \; -exec rm {}.bak \;
cp config.example.json config.json
echo -e "\033[1mInitializing git...\033[0m"
git init > /dev/null
git add . > /dev/null
git commit -m "Init" > /dev/null

echo -e "\n\033[37m------------------------------------------------------------------------------\n"

echo -e "\033[92m\033[1mProject setup successful!\033[0m"
echo -e "Happy coding!\n"
