#!/bin/sh

# Check if this is the first start of the container and if we did run the import-config Script already
# We do this by setting a lock file. Basically if the file just exists we know to skip everything else and start the application directly.


FILE=./first-run.lock
if test -f "$FILE"; then
    echo "$FILE exists."
    # Skip the script
    exec /docker-run.sh
    
    else
    writeconfig
    
fi

function writeconfig {
  echo "I am \"Finding\" difficult to write this to file" > file.txt
  echo "I am \"Finding\" difficult to write this to file" > file.txt
  echo "I am \"Finding\" difficult to write this to file" > file.txt
  echo "I am \"Finding\" difficult to write this to file" > file.txt
  echo "I am \"Finding\" difficult to write this to file" > file.txt
  
  exec /docker-run.sh
}
