{
  description = "Flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    { nixpkgs, ... }:
    let
      forAllSystems = nixpkgs.lib.genAttrs [
        "aarch64-darwin"
        "x86_64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];

      pkgsFor =
        system:
        import nixpkgs {
          inherit system;
          config = {
            allowUnfree = true;
          };
        };

      aws-nuke =
        system:
        let
          pkgs = pkgsFor system;
        in
        pkgs.buildGoModule {
          pname = "aws-nuke";
          version = "2.25.0";
          src = ./.;
          vendorHash = "sha256-nYCRKEPVZxzeuleb/u7FVu1Kd8275RuWy7diA7aYrwo=";
          doCheck = false;
        };

      shell =
        system:
        let
          pkgs = pkgsFor system;
        in
        pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.golangci-lint
          ];
        };
    in
    {
      packages = forAllSystems (system: {
        default = aws-nuke system;
        aws-nuke = aws-nuke system;
      });

      devShells = forAllSystems (system: {
        default = shell system;
      });
    };
}
