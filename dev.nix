# SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
#
# SPDX-License-Identifier: CC0-1.0

# Creates a Nix development environment for the Beacon project.
let
  # Branch: nixos-unstable
  # Date of commit: 2026-03-28
  commit_ref = "8110df5ad7abf5d4c0f6fb0f8f978390e77f9685";
  nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/tarball/${commit_ref}";
  pkgs = import nixpkgs {
    config = { };
    overlays = [ ];
  };
in

pkgs.mkShellNoCC {
  packages = with pkgs; [
    git
    go_1_26
    golangci-lint
    gopls
    reuse
    tmux
  ];

  TMUX_SESSION = "Beacon Development";

  shellHook = ''
    export GOROOT=$( which go | xargs dirname | xargs dirname )/share/go
    tmux new-session -d -s "$TMUX_SESSION"
    tmux send-keys "alias mage=\"go tool -modfile=tools/tools.mod mage\" && clear" C-m
    exec tmux attach -t "$TMUX_SESSION"
  '';
}
