{
  description = "Open the web URL of the current Git repository in a browser";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;

      # Keep in sync with the latest release tag when packaging.
      packageVersion = "2.4.2";
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          commit = self.shortRev or self.dirtyShortRev or "dirty";
        in
        {
          default = self.packages.${system}.git-open;
          git-open = pkgs.buildGoModule {
            pname = "git-open";
            version = packageVersion;
            src = self;

            # Bump this after changing go.mod / go.sum:
            #   nix build 2>&1 | rg 'got:'
            vendorHash = "sha256-WaM79VI/0eSCj4t6Myji/WPrjRSUFfaQjHfqPXGULgM=";

            # Needed by worktree-related tests in checkPhase.
            nativeBuildInputs = [ pkgs.git ];

            ldflags = [
              "-s"
              "-w"
              "-X github.com/zhaochunqi/git-open/cmd.Version=${packageVersion}"
              "-X github.com/zhaochunqi/git-open/cmd.CommitHash=${commit}"
              "-X github.com/zhaochunqi/git-open/cmd.BuildDate=unknown"
            ];

            # ldflags inject release version, so skip tests that assert the
            # default "dev" values from source.
            checkFlags = [ "-skip=^Test_rootCmd_VersionFlag$" ];

            meta = with pkgs.lib; {
              description = "Open the web URL of the Git repository";
              homepage = "https://github.com/zhaochunqi/git-open";
              changelog = "https://github.com/zhaochunqi/git-open/releases/tag/v${packageVersion}";
              license = licenses.mit;
              mainProgram = "git-open";
              platforms = platforms.unix ++ platforms.windows;
            };
          };
        }
      );

      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.git-open}/bin/git-open";
        };
      });

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              gotools
              git
            ];
          };
        }
      );

      formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.nixfmt-rfc-style);
    };
}
