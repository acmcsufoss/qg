{ pkgs ? import <nixpkgs> {} }:

let	goapi-gen = pkgs.buildGoModule {
		name = "goapi-gen";
		version = "081d60b";

		src = pkgs.fetchFromGitHub {
			owner  = "discord-gophers";
			repo   = "goapi-gen";
			rev    = "074e08fcdf14d45fc9b8812fa97c3c05f5f97626";
			sha256 = "109q1vxp8y131n6b0dxvg2zb9vqylxgn6q9n8y2d4v5yhh88436r";
		};

		vendorSha256 = "1dknfg3w97421c8dnld5kvx0psicvmxr7wzkhqipaxplcg3cqrr9";
	};

	sqlc = pkgs.buildGoModule {
		name = "sqlc";
		version = "1.12.0";

		src = pkgs.fetchFromGitHub {
			owner  = "kyleconroy";
			repo   = "sqlc";
			rev    = "45bd150";
			sha256 = "1np2xd9q0aaqfbcv3zcxjrfd1im9xr22g2jz5whywdr1m67a8lv2";
		};

		proxyVendor = true;
		vendorSha256 = "0mk1bs8ppis7dr3mcg73j8abvi1qpbg06adx8sxpzrfag4i9vg8k";
	};

	goose = pkgs.buildGoModule {
		name = "goose";
		version = "3.5.3";

		src = pkgs.fetchFromGitHub {
			owner  = "pressly";
			repo   = "goose";
			rev    = "5f1f43cfb2ba11d901b1ea2f28c88bf2577985cb";
			sha256 = "13hcbn4v78142brqjcjg1q297p4hs28n25y1fkb9i25l5k2bwk7f";
		};

		vendorSha256 = "1yng6dlmr4j8cq2f43jg5nvcaaik4n51y79p5zmqwdzzmpl8jgrv";
		subPackages = [ "cmd/goose" ];
	};

	nixos-shell = pkgs.buildGoModule {
		name = "nixos-shell";
		src = pkgs.fetchFromGitHub {
			owner  = "diamondburned";
			repo   = "nixos-shell";
			rev    = "e238cb522f7168fbc997101d00e6e2cc0d3e2ff9";
			sha256 = "02wqbfmc0c7q3896x6k2hxwcf1x202qfw0almb6rchlh7cqkva0w";
		};
		vendorSha256 = "0gjj1zn29vyx704y91g77zrs770y2rakksnn9dhg8r6na94njh5a";
	};

in pkgs.mkShell {
	buildInputs = with pkgs; [
		go_1_18
		gopls
		goapi-gen
		goose
		sqlc
		pgformatter
		nixos-shell # for local PostgreSQL server
		nodejs
	];

	shellHook = ''
		PATH="$PWD/frontend/node_modules/.bin:$PATH"
	'';
}
