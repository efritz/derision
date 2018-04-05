#!/usr/bin/env python3

import derision
import requests


def main():
    with derision.Client() as d:
        requests.post(
            url='http://localhost:5000/',
            data={'bonk': 'quux'},
            files={
                'foo.key': open('files/foo.txt'),
                'bar.key': open('files/bar.txt'),
                'baz.key': open('files/baz.txt'),
            },
        )

        request = d.get_requests()[0]
        assert request.form == {
            'bonk': ['quux'],
        }

        assert request.files == {
            'foo.txt': 'content #1\n',
            'bar.txt': 'content #2\n',
            'baz.txt': 'content #3\n',
        }

    print('Done.')


if __name__ == '__main__':
    main()
