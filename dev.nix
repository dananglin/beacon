# Creates a development environment for the Enbas project.
let
  # Branch: nixos-unstable
  # Date of commit: 2025-10-07
  commit_ref = "c9b6fb798541223bbb396d287d16f43520250518";
  nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/tarball/${commit_ref}";
  pkgs = import nixpkgs {
    config = { };
    overlays = [ ];
  };
in

pkgs.mkShellNoCC {
  packages = with pkgs; [
    git
    go
    go-grip
    golangci-lint
    gopls
    reuse
    tmux
  ];

  shellHook = ''
    export GOROOT=$( which go | xargs dirname | xargs dirname )/share/go
    tmux new-session -d -s "Beacon Development"
    tmux send-keys "alias mage=\"go tool -modfile=tools/tools.mod mage\" && clear" C-m
    exec tmux attach -t "Beacon Development"
  '';
}
