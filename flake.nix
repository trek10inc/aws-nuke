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
          vendorHash = "sha256-VyZ0DF//oVjlQFjIegZf4ITli3DS1IF99HHYYP18+4E=";
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
