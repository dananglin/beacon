# SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
#
# SPDX-License-Identifier: CC0-1.0

# Creates a Nix development environment for the Beacon project.
let
  # Branch: nixos-unstable
  # Date of commit: 2026-03-02
  commit_ref = "cf59864ef8aa2e178cccedbe2c178185b0365705";
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
    mdbook
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
