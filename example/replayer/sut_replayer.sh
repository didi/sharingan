#!/bin/bash

workspace=$(cd $(dirname $0) && pwd -P)
cd ${workspace}

# TODO: Attention please!!! Modify-1! Your project's binary name. Is better that $app not contain '-'。
app=example
appcov="$app.test"
# TODO: Attenton please!!! Modify-2! Your project's root path.
root=$(dirname $(dirname $workspace))
# TODO: Attenton please!!! Modify-3! The path where you git clone sharingan. It's used to copy main_test.go & gocov tool when failed by wgetting from github.
replayer_agent_root=$root/replayer-agent

# set file
sut_file=$workspace/$app
sut_file_cov=$workspace/$appcov
sut_pid_file=$workspace/sut.pid
sut_log_file=$workspace/sut.log

# set msg
info_msg="\033[;32m[INFO]\033[0m\t"
warn_msg="\033[;33m[WARN]\033[0m\t"
error_msg="\033[;31m[ERROR]\033[0m\t"

#cov
cov_cmd=cov
cov_file_path='/tmp/ShaRinGan/'
cov_file_prefix='coverage.'
cov_file_suffix='.cov'

# set web server
local_ip="127.0.0.1"
local_port="8998"

# set github about go
git_replayer_agent='https://github.com/didi/sharingan/raw/master/replayer-agent'
git_go='https://raw.githubusercontent.com/didi/sharingan-go/recorder'
install_go_md='https://github.com/didi/sharingan-go/tree/recorder'
VERSION="go1.13"

function install_go() {
        curl "$git_go/install/$VERSION" >> /dev/null
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at installing sharingan-go $VERSION : 'curl $git_go/install/$VERSION' \n"
            printf "${error_msg}Please install sharingan-go $VERSION manually (refer to $install_go_md) and try again!!!  \n"
            printf "${warn_msg}Please execute 'export GOROOT=/tmp/recorder-$VERSION && export PATH=\$GOROOT/bin:\$PATH' after installing $VERSION!!!\n"
            exit 1
        fi
        curl "$git_go/install/$VERSION" | sh
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at installing sharingan-go $VERSION : 'curl $git_go/install/$VERSION | sh' \n"
            printf "${error_msg}Please install sharingan-go $VERSION manually (refer to $install_go_md) and try again!!!  \n"
            printf "${warn_msg}Please execute 'export GOROOT=/tmp/recorder-$VERSION && export PATH=\$GOROOT/bin:\$PATH' after installing $VERSION!!!\n"
            exit 1
        fi
        export GOROOT="/tmp/recorder-$VERSION"
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at setting GOROOT, please check!!!\n"
            exit 1
        fi
        export PATH=$GOROOT/bin:$PATH
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at updating PATH, please check!!!\n"
            exit 1
        fi

        printf "${info_msg}GOROOT=$GOROOT \n"
}

# install go
function install_go_tar() {
        # get go url
        releases_go='https://github.com/didi/sharingan-go/releases/download/'
        if [[ "$OSTYPE" =~ ^darwin.* ]]; then
            goSuffix="$VERSION.recorder/$VERSION.darwin-amd64.tar.gz"
        elif [[ "$OSTYPE" =~ ^linux.* ]]; then
            goSuffix="$VERSION.recorder/$VERSION.linux-amd64.tar.gz"
        else
            printf "${error_msg} Unknown system type! build failed at installing go! \n"
            exit 1
        fi

        # install golang
        recorder_dir=/tmp/recorder-${VERSION}
        wget "$releases_go/$goSuffix" -O "$recorder_dir.tar.gz"
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at downloading sharingan-go $VERSION : wget $releases_go/$goSuffix -O $recorder_dir.tar.gz !!!\n"
            printf "${error_msg}Please install sharingan-go $VERSION manually (refer to $install_go_md) and try again!!!  \n"
            printf "${warn_msg}Please execute 'export GOROOT=/tmp/recorder-$VERSION && export PATH=\$GOROOT/bin:\$PATH' after installing $VERSION!!!\n"
            exit 1
        fi
        rm -rf $recorder_dir && mkdir -p $recorder_dir
        tar -xzf "$recorder_dir.tar.gz" -C ${recorder_dir} --strip-components=1
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at installing $VERSION, please check!!!\n"
            exit 1
        fi
        rm -rf "$recorder_dir.tar.gz"

        # export
        export GOROOT="/tmp/recorder-$VERSION"
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at setting GOROOT, please check!!!\n"
            exit 1
        fi
        export PATH=$GOROOT/bin:$PATH
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at updating PATH, please check!!!\n"
            exit 1
        fi

        printf "${info_msg}GOROOT=$GOROOT \n"
}

# build
function build() {
    preGoroot="/tmp/recorder-go"
    if [[ -z $GOROOT || $GOROOT != $preGoroot* ]];then
        printf "${info_msg}Installing sharingan-go for you ~~~ \n"
        install_go
    fi

    which go &> /dev/null
    if [ $? -ne 0 ]; then
        printf "${error_msg}build failed at executing 'which go', please execute 'export PATH=\$GOROOT/bin:\$PATH' !!!\n"
        exit 1
    fi
    goVer=`go version`
    if [[ $goVer < 'go version go1.11' ]];then
        printf "${warn_msg}current is $goVer, go mod requires at least go1.11!!!\n"
        install_go
    fi

    if [ -z $GOPATH  ];then
        mkdir -p /tmp/replayer
        export GOPATH=/tmp/replayer
        printf "${warn_msg}No GOPATH, so setting GOPATH=/tmp/replayer \n"
    fi

    cd $root
    go get 'github.com/didi/sharingan'
    if [ $? -ne 0 ]; then
        printf "${error_msg}build failed at executing go get github.com/didi/sharingan, please check!!!\n"
        exit 1
    fi
    go clean -modcache
    printf "${info_msg}go mod download, please wait~ \n"
    go mod download

    cd $workspace
    binName=$app
    if [[ $sub_cmd == $cov_cmd ]];then
        # wget main_test.go
        wget -O main_test.go "$git_replayer_agent/install/codeCov/main_test.go"
        if [ $? -ne 0 ]; then
            printf "${warn_msg}Failed to wget main_test.go. Now trying another way, copying from sharingan/replayer-agent!!!\n"
            cp -rf $replayer_agent_root/install/codeCov/main_test.go .
            if [ $? -ne 0 ]; then
                printf "${error_msg}build failed at copying main_test.go from $replayer_agent_root/install/codeCov/, please update var \$replayer_agent_root!!!\n"
                exit 1
            fi
            printf "${info_msg}Successed to copy main_test.go. \n"
        fi

        # wget gocov* tool
        if [[ "$OSTYPE" =~ ^darwin.* ]]; then
            sys_dir='darwin'
            wget -O gocov "$git_replayer_agent/install/codeCov/darwin/gocov"
            wget -O gocov-html "$git_replayer_agent/install/codeCov/darwin/gocov-html"
        elif [[ "$OSTYPE" =~ ^linux.* ]]; then
            sys_dir='linux'
            wget -O gocov "$git_replayer_agent/install/codeCov/linux/gocov"
            wget -O gocov-html "$git_replayer_agent/install/codeCov/linux/gocov-html"
        else
            printf "${error_msg} Unknown system type! build failed at wgetting gocov* tool! \n"
            exit 1
        fi
        if [ $? -ne 0 ]; then
            printf "${warn_msg}Failed to wget gocov* tool, Now trying another way, copying from sharingan/replayer-agent!!!\n"
            cp -rf $replayer_agent_root/install/codeCov/$sys_dir/gocov* .
            if [ $? -ne 0 ]; then
                printf "${error_msg}build failed at copying gocov* tool from $replayer_agent_root/install/codeCov/$sys_dir/, please update var \$replayer_agent_root!!!\n"
                exit 1
            fi
            printf "${info_msg}Successed to copy gocov* tool. \n"
        fi

        chmod 755 $workspace/gocov*
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at chmod $workspace/gocov* tool, please check!!!\n"
            exit 1
        fi

        binName=$appcov
    fi

    pkgReplay='_ "github.com/didi/sharingan"'
    find ./ -name "*\.go" -maxdepth 1 | xargs -n 1 grep  $pkgReplay &> /dev/null
    if [ $? -eq 0 ];then
        printf "${error_msg} build failed for not importing package sharingan/replayer, please import it!!!\n"
        exit 1
    fi

    if [[ $sub_cmd == $cov_cmd ]];then
        go test -o $appcov -gcflags="all=-N -l" -tags="replayer" -v -c -covermode=count -coverpkg ./...
    else
        go build -o $app -tags="replayer" -gcflags="all=-N -l"
    fi
    if [ $? -ne 0 ]; then
        printf "${error_msg}build failed at building sut, please check!!!\n"
        exit 1
    fi

    printf "${info_msg}SUT $binName builds success!!!\n"
}

# start
function start() {
    binName=$app
    binFile=$sut_file
    cmdStart="$sut_file >> $sut_log_file"
    cmdBuild="sh sut_replayer.sh build"
    if [[ $sub_cmd == $cov_cmd ]];then
        binName=$appcov
        binFile=$sut_file_cov
        curTime=`date +%s`
        cmdStart="SYSTEM_TEST=true $sut_file_cov -test.coverprofile=$cov_file_path$cov_file_prefix$app.$curTime$cov_file_suffix >> $sut_log_file"
        cmdBuild="sh sut_replayer.sh build $cov_cmd"
    fi

    if [ ! -f "$binFile" ]; then
        printf "${error_msg}please exesute $cmdBuild first!!!\n"
        exit 1
    fi

    if [[ "$OSTYPE" =~ ^linux.* ]]; then
        local_ip=`ip addr | grep "inet " | grep -v 127 | awk '{print $2}' | awk -F '/' '{print $1}'`
    fi

    #ps -ef | grep $binName | grep -v grep $grepV &> /dev/null
    if [[ $sub_cmd == $cov_cmd ]];then
        ps -ef | grep $binName | grep -v grep  &> /dev/null
    else
        ps -ef | grep $binName | grep -v grep | grep -v $appcov  &> /dev/null
    fi
    if [ $? -eq 0 ];then
        printf "${warn_msg}SUT $binName has already started!!! Have fun with http://${local_ip}:${local_port} !\n"
        exit 0
    fi

    # 默认值，如有需要可以修改
    export REPLAYER_MOCK_IP="127.0.0.1"
    export REPLAYER_MOCK_PORT="3515"

    nohup $cmdStart >> $sut_log_file 2>&1 &
    sleep 2
    pid=$!

    if [[ $sub_cmd == $cov_cmd ]];then
        ps -ef | grep $binName | grep -v grep  &> /dev/null
    else
        ps -ef | grep $binName | grep -v grep | grep -v $appcov  &> /dev/null
    fi
    if [ $? -ne 0 ];then
        printf "${error_msg} Failed to start sut $binName, please check $sut_log_file and try again!!!\n"
        exit 1
    fi
    echo $pid > $sut_pid_file
    printf "${info_msg}SUT $binName starts success!!! [pid=${pid}]  Have fun with http://${local_ip}:${local_port} !\n"
}

# stop
function stop() {
    binName=$app
    if [[ $sub_cmd == $cov_cmd ]];then
        binName=$appcov
    fi

    if [[ $sub_cmd == $cov_cmd ]];then
        ps -ef | grep $binName | grep -v grep  &> /dev/null
    else
        ps -ef | grep $binName | grep -v grep | grep -v $appcov  &> /dev/null
    fi
    if [ $? -ne 0 ];then
        printf "${warn_msg}SUT $binName is not running!!!\n"
        exit 0
    fi

    pkill $binName
    sleep 2

    if [[ $sub_cmd == $cov_cmd ]];then
        ps -ef | grep $binName | grep -v grep  &> /dev/null
    else
        ps -ef | grep $binName | grep -v grep | grep -v $appcov  &> /dev/null
    fi
    if [ $? -eq 0 ];then
        printf "${error_msg}Failed to stop sut $binName, please check and try again!!!\n"
        exit 1
    fi
    printf "${info_msg}SUT $binName stops success!!!\n"

    # codeCov report
    if [[ $sub_cmd == $cov_cmd ]];then
        f="$cov_file_path$cov_file_prefix"'*'"$app"'*'"$cov_file_suffix"
        covFile=`ls -trls $f | tail -n 1`
        #echo $covFile
        if [[ $covFile != *$cov_file_suffix ]];then
            printf "${warn_msg}Failed to get *.cov file, please check $cov_file_path!!!\n"
            exit 0
        fi

        # get the full path of covfile created by stop cmd
        covFile=/`echo ${covFile#*/}`
        curTime=`date +%s`
        covFile_renamed=`echo ${covFile%cov*}`$curTime
        mv $covFile $covFile_renamed
        if [ $? -ne 0 ];then
            printf "${warn_msg}Failed to rename $covFile to $covFile_renamed, please check !!!\n"
            exit 0
        fi

        # convert cov file to html
        covNameOnly_html=`echo ${covFile_renamed##*/}`
        $workspace/gocov convert $covFile_renamed | $workspace/gocov-html > $workspace/$covNameOnly_html'.html'
        mv $workspace/$covNameOnly_html'.html' $cov_file_path/$covNameOnly_html'.html'
        if [ $? -ne 0 ];then
            printf "${warn_msg}Failed to gocov convert $covFile_renamed to $covFile_renamed'.html', please check !!!\n"
            exit 0
        fi

        if [[ "$OSTYPE" =~ ^linux.* ]]; then
            local_ip=`ip addr | grep "inet " | grep -v 127 | awk '{print $2}' | awk -F '/' '{print $1}'`
        fi

        printf "${info_msg}Have fun with http://${local_ip}:${local_port}/coverage/report/$covNameOnly_html.html !!!\n"
    fi
}

# cov or null
sub_cmd=$2
# check sub cmd
if [[ ! -z $sub_cmd && $sub_cmd != $cov_cmd ]];then
    printf "${warn_msg} cmd lists:\n1. build [$cov_cmd]\n2. start [$cov_cmd]\n3. stop  [$cov_cmd]\n4. reload  [$cov_cmd]\n[$cov_cmd] is optional for codeCov\n"
    exit 1
fi

case $1 in
    "build" )
        build
        ;;
    "start" )
        start
        ;;
    "stop" )
        stop
        ;;
    "reload" )
        stop
        start
        ;;
    * )
        printf "${warn_msg} cmd lists:\n1. build [$cov_cmd]\n2. start [$cov_cmd]\n3. stop  [$cov_cmd]\n4. reload  [$cov_cmd]\n[$cov_cmd] is optional for codeCov\n"
        exit 1
        ;;
esac
