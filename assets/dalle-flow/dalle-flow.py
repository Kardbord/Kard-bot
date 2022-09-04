#!/usr/bin/env python3

import sys
import argparse
import subprocess


def main():
    server_url = 'grpcs://dalle-flow.dev.jina.ai'
    output_file = 'output.png'
    parser = argparse.ArgumentParser(
        description='A simple script to generate AI art from a prompt utilizing dalle-flow, and save the resulting image to disk.')
    parser.add_argument('-p', '--prompt', type=str,
                        help='The prompt to pass to Dalle-Flow.', required=True)
    parser.add_argument('-s', '--server', type=str,
                        help='The Dalle-Flow server to connect to. Default: {}.'.format(server_url), default=server_url)
    parser.add_argument('-o', '--output', type=str,
                        help='The filename to write the image to. Default: {}.'.format(output_file), default=output_file)
    args = parser.parse_args()

    try:
        from docarray import Document
    except:
        subprocess.check_call(
            [sys.executable, '-m', 'pip', 'install', '"docarray[common]>=0.13.5"', 'jina'])
    finally:
        from docarray import Document

    prompt = args.prompt
    server_url = args.server
    output_file = args.output
    doc = Document(text=prompt).post(server_url, parameters={'num_images': 1})
    img = doc.matches[0]
    img.embedding = doc.embedding
    diffused = img.post(f'{server_url}', parameters={
                        'skip_rate': 0.5, 'num_images': 1}, target_executor='diffusion').matches[0]
    diffused = diffused.post(f'{server_url}/upscale')
    diffused.save_uri_to_file(output_file)
    return 0


if __name__ == '__main__':
    try:
        sys.exit(main())
    except DeprecationWarning:
        sys.exit(0)
    except Exception as e:
        print(e)
        sys.exit(1)

