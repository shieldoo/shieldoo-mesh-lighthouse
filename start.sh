#/bin/sh

# catch terminate signals
trap 'exit 0' SIGTERM

echo '
       .__    .__       .__       .___             
  _____|  |__ |__| ____ |  |    __| _/____   ____  
 /  ___/  |  \|  |/ __ \|  |   / __ |/  _ \ /  _ \ 
 \___ \|   Y  \  \  ___/|  |__/ /_/ (  <_> |  <_> )
/____  >___|  /__|\___  >____/\____ |\____/ \____/ 
     \/     \/        \/           \/             '
echo ''

if [[ -n "$MYCONFIG" ]]; then
    # create config 
    echo ''
    echo 'creating config ..'
    echo $MYCONFIG | base64 -d > /lib/config/myconfig.yaml
fi

mkdir -p /dev/net
if [ ! -c /dev/net/tun ]; then
    mknod /dev/net/tun c 10 200
fi

echo ''
echo 'starting lighthouse ..'

/app/shieldoo-mesh-lighthouse