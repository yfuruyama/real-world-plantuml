#!/usr/bin/env perl
use strict;
use warnings;

while (<STDIN>) {
    chomp $_;
    my $web_url = $_;
    if (my ($owner, $repo, $ref, $path) = ($web_url =~ m!^https://github.com/([^/]+)/([^/]+)/blob/([^/]+)/(.+)$!)) {
        print sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s\n", $owner, $repo, $path, $ref);
    }
}
