{
  description = "Seanime - self-hosted anime/manga media server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs = inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      perSystem = { pkgs, lib, ... }:
        let
          version = "3.5.2";

          seanimeWeb = pkgs.buildNpmPackage {
            pname = "seanime-web";
            inherit version;
            src = ./seanime-web;
            npmDepsHash = "sha256-fWlK2h0RQF9GnEogXW3bwM01RCCDVij/9S2sn2BA3S4=";
            buildPhase = "npm run build";
            installPhase = "cp -r out $out";
          };

          seanime = pkgs.buildGoModule {
            pname = "seanime";
            inherit version;
            src = ./.;
            vendorHash = "sha256-TN9shH4B7XVDIa541+7MHTNQs1IKPRJW1dn8tmES5jg=";
            subPackages = [ "." ];
            env.CGO_ENABLED = 1;
            nativeBuildInputs = with pkgs; [ gcc pkg-config ]
              ++ lib.optionals pkgs.stdenv.isDarwin [
                darwin.apple_sdk.frameworks.Security
                darwin.apple_sdk.frameworks.CoreFoundation
              ];
            buildInputs = [ pkgs.sqlite ];
            preBuild = "mkdir -p web && cp -r ${seanimeWeb}/* web/";
            ldflags = [ "-s" "-w" ];
            meta = {
              description = "Self-hosted media server for anime and manga";
              homepage = "https://seanime.app";
              license = pkgs.lib.licenses.gpl3Only;
              mainProgram = "seanime";
              platforms = pkgs.lib.platforms.unix;
            };
          };

          seanime-denshi = pkgs.appimageTools.wrapType2 {
            pname = "seanime-denshi";
            inherit version;
            src = pkgs.fetchurl {
              url = "https://github.com/5rahim/seanime/releases/download/v${version}/seanime-denshi-${version}_Linux_x86_64.AppImage";
              hash = "sha256-8erYkDgOE5Ma4X502JEyXpYnfrLagJA0i0ePuRE+N4s=";
            };
            meta = {
              description = "Seanime Denshi desktop client";
              homepage = "https://seanime.app";
              license = pkgs.lib.licenses.gpl3Only;
              mainProgram = "seanime-denshi";
              platforms = [ "x86_64-linux" ];
            };
          };
        in
        {
          packages = {
            inherit seanime seanimeWeb seanime-denshi;
            default = seanime-denshi;
          };

          devShells.default = pkgs.mkShell {
            buildInputs = with pkgs; [ go_1_23 gopls gotools nodejs_20 sqlite gcc pkg-config ];
          };

          formatter = pkgs.nixfmt-rfc-style;
        };

      flake.nixosModules.default = { config, lib, pkgs, ... }:
        let cfg = config.services.seanime;
            pkg = inputs.self.packages.${pkgs.system}.seanime;
        in {
          options.services.seanime = {
            enable       = lib.mkEnableOption "Seanime media server";
            dataDir      = lib.mkOption { type = lib.types.str; default = "/var/lib/seanime"; };
            user         = lib.mkOption { type = lib.types.str; default = "seanime"; };
            group        = lib.mkOption { type = lib.types.str; default = "seanime"; };
            openFirewall = lib.mkOption { type = lib.types.bool; default = false; };
          };

          config = lib.mkIf cfg.enable {
            users.users.${cfg.user}  = { isSystemUser = true; group = cfg.group; home = cfg.dataDir; createHome = true; };
            users.groups.${cfg.group} = {};

            systemd.services.seanime = {
              description = "Seanime Media Server";
              wantedBy = [ "multi-user.target" ];
              after    = [ "network.target" ];
              serviceConfig = {
                ExecStart       = "${pkg}/bin/seanime --datadir=${cfg.dataDir}";
                User            = cfg.user;
                Group           = cfg.group;
                Restart         = "on-failure";
                RestartSec      = "5s";
                NoNewPrivileges = true;
                PrivateTmp      = true;
                ProtectSystem   = "strict";
                ReadWritePaths  = [ cfg.dataDir ];
              };
            };

            networking.firewall.allowedTCPPorts = lib.mkIf cfg.openFirewall [ 43211 ];
          };
        };
    };
}
