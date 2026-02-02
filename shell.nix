with import <nixpkgs> {};

stdenv.mkDerivation {
name = "go-env";

buildInputs = [
		go
		syft
		grype
		docker
		docker-credential-helpers
		trivy
];

SOURCE_DATE_EPOCH = 315532800;
PROJDIR = "${toString ./.}";
S_NETWORK="weave";
S_HOSTNAME = "m3sframework.weave.local";
S_VOLUME_RW_1 = "/var/run/docker.sock";

shellHook = ''
		export LD_LIBRARY_PATH="${pkgs.stdenv.cc.cc.lib}/lib"
		export PATH=/tmp/bin:$PATH
		export GOTMPDIR=/tmp
		export TMPDIR=/tmp
		mkdir /tmp/bin
		'';
}
