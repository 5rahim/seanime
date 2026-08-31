{
  appimage-run,
  fetchurl,
  lib,
  stdenv,

  bashNonInteractive,
  ...
}:
let
  pname = "seanime-denshi";
  version = "3.10.2";
  src = fetchurl {
    url = "${meta.downloadPage}/download/v${version}/${pname}-${version}_Linux_x86_64.AppImage";
    hash = "sha256-Appt2gh4mYyz1YMY4uvNmpXGkKVqxDimABNKMhTbZMA=";
  };
  meta = {
    description = "An anime+manga browser|downloader|viewer";
    homepage = "https://seanime.app/";
    downloadPage = "https://github.com/5rahim/seanime/releases";
    license = lib.licenses.gpl3;
    # sourceProvenance = with lib.sourceTypes; [ binaryNativeCode ];
    # maintainers = with lib.maintainers; [ onny ];
    platforms = [ "x86_64-linux" ];
    mainProgram = pname;
  };

  out_exec = "$out/bin/${pname}";
  desktop = /* ini */ ''
    [Desktop Entry]
    Type=Applicaiton
    Exec=${out_exec}
    Icon=seanime
    Terminal=false

    Name=SeaAnime Denshi
    Comment=${meta.description}

    StartupNotify=true
    Categories=Network;Player;TV;Video;
    Keywords=${pname};seanime;denshi;anime;manga;torrent;
  '';
  launcher-script = /* sh */ ''
    #!${lib.getExe bashNonInteractive}
    ${lib.getExe appimage-run} $src
  '';
in
stdenv.mkDerivation {
  inherit
    pname
    version
    src
    meta
    ;

  buildInputs = [ appimage-run ];
  dontUnpack = true;

  # makeWrapper ${a.lib.getExe appimage-run} $out/bin/${pname} \
  # --add-flags "$out/bin/${pname}.AppImage"

  installPhase = ''
    mkdir -p $out/{bin,share/applications}
    cp $src $out/bin/${pname}.AppImage

    cat > ./${pname}.sh << EOF
    ${launcher-script}
    EOF
    cat > ./${pname}.desktop << EOF
    ${desktop}
    EOF

    install -Dm 755 ${pname}.sh ${out_exec}
    install -Dm 644 ${pname}.desktop $out/share/applications/${pname}.desktop

    chmod +x $out/bin/${pname}
  '';

  # extraInstallCommands = ''
  #   substituteInPlace $out/share/applications/${pname}.desktop \
  #     --replace-fail 'Exec=AppRun' 'Exec=${meta.mainProgram}'
  # '';
}
