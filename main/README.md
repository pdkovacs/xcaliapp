# XCali Application

## Installation for develoment

1. Clone the required modules into `~/workspace/xcaliapp`

    ```bash
    mkdir -p ~/workspace/xcaliapp
    cd ~/workspace/xcaliapp
    for dir in main server webclient lambda s3store gitstore;
    do      
        echo $dir
        if [ -d "$dir" ]; then
            echo "Directory '$dir' already exist"
            continue
        fi
        git clone git@github.com:xcaliapp/$dir.git
    done 
    ```
1. Set up the go workspace:
    ```bash
    cd main
    task setup
    ```
