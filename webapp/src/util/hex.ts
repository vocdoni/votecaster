
export const toArrayBuffer = (input: string): Uint8Array => {
  if (input.length % 2 !== 0) {
    input = input + '0';
  }

  const view = new Uint8Array(input.length / 2);
  for (let i = 0; i < input.length; i += 2) {
    view[i / 2] = parseInt(input.substring(i, i + 2), 16);
  }
  return view;
}