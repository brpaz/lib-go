{
  pkgs,
  ...
}:
{

  packages = with pkgs; [
    go-task
    lefthook
    (gomarkdoc.overrideAttrs (_: { doCheck = false; }))
    golangci-lint
    commitlint-rs
  ];

  languages = {
    go = {
      enable = true;
      package = pkgs.go_1_26;
      lsp = {
        package = pkgs.gopls;
      };
    };
  };

  enterShell = ''
    go version
    lefthook install
  '';

  # https://devenv.sh/tests/
  enterTest = ''
    echo "Running tests"
    go test ./... -v
  '';
}
