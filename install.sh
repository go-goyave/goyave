#!/bin/bash

set -e

usage() { echo "Usage: $0 <project_name>" 1>&2; exit 1; }

if [ $# -ne 1 ]; then
    usage
fi

if [ -d $1 ]; then
  echo -e "\e[31m\e[1mError: \e[0mdirectory \"\e[37m$1\e[0m\" already exists."
  exit 1
fi

echo -e "\e[36m\e[1m                                                             
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
\e[0m"
echo -e "\e[37m------------------------------------------------------------------------------\n"

echo -e "\e[92m\e[1mThank you for using Goyave!\e[0m"
echo -e "If you like the framework, please consider supporting me on Patreon:\n\e[37mhttps://www.patreon.com/bePatron?u=25997573\e[0m\n"

echo -e "\e[37m------------------------------------------------------------------------------\n"

echo -e "\e[1mDownloading template project...\e[0m"
curl -LOk https://github.com/System-Glitch/goyave-template/archive/master.zip

echo -e "\e[1mUnzipping...\e[0m"
unzip -q master.zip
rm master.zip
echo -e "\e[1mSetup...\e[0m"
mv goyave-template-master $1
cd $1
find ./ -type f \( -iname \*.go -o -iname \*.mod -o -iname \*.json \) -exec sed -i "s/goyave_template/$1/g" {} \;
cp config.example.json config.json
echo -e "\e[1mInitializing git...\e[0m"
git init > /dev/null
git add . > /dev/null
git commit -m "Init" > /dev/null

echo -e "\n\e[37m------------------------------------------------------------------------------\n"

echo -e "\e[92m\e[1mProject setup successful!\e[0m"
echo -e "Happy coding!\n"