{ pkgs, self}:

  pkgs.buildGoModule {
  pname = "zcli";
  version = "0..1";

  src = self; #zcli;

  nativeBuildInputs = with pkgs; [ go ];
  vendorHash = null;

  installPhase = ''
  mkdir -p $out/bin
  cp $GOPATH/bin/zcli $out/bin/
  echo "Installed zcli to $out/bin/"
  ls -l $out/bin
    '';

  meta = with pkgs.lib; {
    description = "A command-line interface (CLI) tool built with Go";
    homepage = "https://github.com/zeropsio/zcli";
    license = licenses.mit;
    maintainers = with maintainers; [ arbel-arad nermalcat69 ];
  };
}
