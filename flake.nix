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
          vendorHash = "sha256-S0prBc6QfTrGDKVs/4QrCVONL9xIA/9j5pDJn+ojzgk="; # replace after first build
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

      nixosModules.default = { config, lib, pkgs, ... }: {
        options.services.sharktask-server = {
          enable = lib.mkEnableOption "sharktasks server";
          port = lib.mkOption { type = lib.types.port; default = 8080; };
        };
        config = lib.mkIf config.services.todo-server.enable {
          systemd.services.todo-server = {
            wantedBy = [ "multi-user.target" ];
            serviceConfig = {
              ExecStart = "${self.packages.${pkgs.system}.sharktasks-server}/bin/sharktasks-server";
              StateDirectory = "sharktasks"; # /var/lib/todo, put your sqlite db here
              DynamicUser = true;
            };
          };
        };
      };
    };
}
