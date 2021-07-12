#!/bin/sh
set -x
set -e

java8_version=0
do_check_install() {
	if command_exists java; then
        java8_version="$(java -version 2>&1 | awk -F '"' '/version/ {print $2}' | grep 1.8 | cut -d'_' -f 2)"
	fi
	if [ "$java8_version" -ge "181" ]; then
		echo "Java $java8_version already installed."
		exit 0
	fi
}

do_install_java() {
	echo "# Installing java 8..."
	echo ""
	os_arch=$(getconf LONG_BIT)
	is_arm=""
	if [ "$lsb_dist" = "raspbian" ] || [ "$(uname -m)" = "armv7l" ] || [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "armv8" ]; then
		is_arm="-arm"
	fi
	cd /opt/
	$sh_c 'wget --no-check-certificate '"https://storage.googleapis.com/edgeworx/downloads/jdk/jdk-8u211$is_arm-$os_arch.tar.gz"''
	$sh_c "tar xzf jdk-8u211$is_arm-$os_arch.tar.gz"
	cd /opt/jdk1.8.0_211/	
	case "$lsb_dist" in
		debian|raspbian|ubuntu|mendel)
			$sh_c "update-alternatives --install /usr/bin/java java /opt/jdk1.8.0_211/bin/java 1100"
			;;		
		fedora|centos)
			$sh_c "alternatives --install /usr/bin/java java /opt/jdk1.8.0_211/bin/java 4"
			;;
	esac
}

do_install_deps() {
	local installer=""
	case "$lsb_dist" in
		ubuntu|debian|raspbian)
			installer="apt"
			;;
		fedora|centos)
			installer="yum"
			;;
	esac

	local iter=0
	while ! $sh_c "$installer update" && [ "$iter" -lt 6 ]; do
		sleep 5
		iter=$((iter+1))
	done

	if [ -z $(command -v wget) ]; then
		$sh_c "$installer install -y wget"
	fi
}

. /etc/iofog/agent/init.sh
init
do_check_install
do_install_deps
do_install_java