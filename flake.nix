{
  description = "shark tasks";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAll = f: nixpkgs.lib.genAttrs systems (s: f nixpkgs.legacyPackages.${s});
    in {
      packages = forAll (pkgs: rec {
        sharktasks-server = pkgs.buildGoModule {
          pname = "sharktasks-server";
          version = "0.1.0";
          src = ./.;
          subPackages = [ "cmd/server" ];
          vendorHash = "sha256-S0prBc6QfTrGDKVs/4QrCVONL9xIA/9j5pDJn+ojzgk=";
          postInstall = "mv $out/bin/server $out/bin/sharktasks-server";
        };
        sharktasks = pkgs.buildGoModule {
          pname = "sharktasks";
          version = "0.1.0";
          src = ./.;
          subPackages = [ "cmd/client" ];
          vendorHash = sharktasks-server.vendorHash;
          postInstall = "mv $out/bin/client $out/bin/sharktasks";
        };
        default = sharktasks;
      });

      devShells = forAll (pkgs: {
        default = pkgs.mkShell { packages = [ pkgs.go pkgs.gopls pkgs.sqlite ]; };
      });

      overlays.default = final: prev: {
        sharktasks-server = self.packages.${final.system}.sharktasks-server;
        sharktasks = self.packages.${final.system}.sharktasks;
      };

      nixosModules.default = { config, lib, pkgs, ... }: 
        let
          yaml = pkgs.formats.yaml {};
          sys = pkgs.stdenv.hostPlatform.system;
          sCfg = config.services.sharktasks-server;
          cCfg = config.programs.sharktasks;
          serverCfgFile = yaml.generate "comfig.yaml" sCfg.settings;
        in 
        {
        options.services.sharktasks-server = {
          enable = lib.mkEnableOption "sharktasks server";
          settings = {
            type = yaml.type;
            default = { 
              port = 8080;
              db = "/var/lib/sharktasks-server/tasks.db";
            };
          };
        };
        options.programs.sharktasks = {
          enable = lib.mkEnableOption "sharktasks";
          settings = lib.mkOption {
            type = yaml.type;
            default = {
              host = "http://localhost";
              port = 8080;
            };
          };
        };
      config = lib.mkMerge [
            (lib.mkIf sCfg.enable {
              environment.etc."sharktasks/config.yaml".source = serverCfgFile;
              systemd.services.sharktasks-server = {
                wantedBy = ["multi-user.target"];
                restartTriggers = [serverCfgFile];
                serviceConfig = {
                  ExecStart = "${self.packages.${sys}.sharktasks-server}/bin/sharktasks-server";
                  StateDirectory = "sharktasks-server";
                  DynamicUser = true;
                };
              };
            })
            (lib.mkIf cCfg.enable {
              environment.systemPackages = [self.packages.${sys}.sharktasks ];
              environment.etc."sharktasks/config.yaml".source = yaml.generate "config.yaml" cCfg.settings;
            })
          ];
    };
  };
}
