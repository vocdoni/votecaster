export const FarcasterLogo = ({ height, fill }: { height?: number; fill: 'purple' | 'white' }) => {
  const h = height || 20
  const width = h * 1.1

  return (
    <svg xmlns='http://www.w3.org/2000/svg' width={width} height={h} fill='none'>
      <title>Farcaster logo</title>
      <g fill={fillColor[fill]} clipPath='url(#a)'>
        <path d='M3.786.05h14.156v2.824h4.025l-.844 2.825h-.714v11.427c.358 0 .65.287.65.642v.77h.13c.358 0 .649.288.649.642v.77h-7.273v-.77c0-.354.29-.642.65-.642h.13v-.77c0-.309.22-.566.512-.628l-.014-6.306c-.23-2.519-2.37-4.493-4.98-4.493-2.608 0-4.75 1.974-4.979 4.494l-.013 6.3c.346.05.772.315.772.633v.77h.13c.358 0 .65.288.65.642v.77H.15v-.77c0-.354.29-.642.649-.642h.13v-.77c0-.355.29-.642.65-.642V5.7H.863L.02 2.874h3.766V.05Z' />
        <path d='M17.942.05h.047V.003h-.047V.05ZM3.786.05V.003H3.74V.05h.047Zm14.156 2.824h-.048v.047h.048v-.047Zm4.025 0 .046.013.018-.06h-.064v.047Zm-.844 2.824v.047h.035l.01-.033-.045-.014Zm-.714 0v-.046h-.047v.046h.047Zm0 11.428h-.047v.047h.047v-.047Zm.65 1.412h-.048v.047h.047v-.047Zm.779 1.412v.047h.046v-.047h-.046Zm-7.273 0h-.048v.047h.048v-.047Zm.78-1.412v.047h.047v-.047h-.048Zm.512-1.398.01.045.038-.007v-.038h-.048Zm-.014-6.306h.048v-.004l-.048.004Zm-9.959 0-.047-.004v.004h.047Zm-.013 6.3h-.048v.041l.04.006.008-.047Zm.772 1.404h-.047v.047h.047v-.047Zm.78 1.412v.047h.047v-.047h-.048Zm-7.273 0H.102v.047H.15v-.047Zm.779-1.412v.047h.047v-.047H.93Zm.65-1.412v.047h.047v-.047h-.047Zm0-11.428h.047v-.046h-.047v.046Zm-.715 0-.045.014.01.033h.035v-.047ZM.02 2.874v-.047h-.063l.018.06.045-.013Zm3.766 0v.047h.048v-.047h-.048ZM17.942.003H3.786v.093h14.156V.003Zm.047 2.87V.05h-.095v2.824h.095Zm3.978-.046h-4.025v.094h4.025v-.094Zm-.798 2.885.844-2.825-.091-.026-.845 2.824.092.027Zm-.76.033h.714v-.093h-.714v.093Zm.048 11.381V5.698h-.095v11.428h.095Zm.649.641a.693.693 0 0 0-.697-.688v.094c.332 0 .602.266.602.595h.095Zm0 .77v-.77h-.095v.77h.095Zm.082-.045h-.13v.093h.13v-.093Zm.697.688a.692.692 0 0 0-.697-.688v.093c.333 0 .602.267.602.595h.095Zm0 .77v-.77h-.095v.77h.095Zm-3.943.047h3.896v-.094h-3.896v.094Zm-2.079 0h2.079v-.094h-2.079v.094Zm-1.298 0h1.298v-.094h-1.298v.094Zm-.048-.817v.77h.095v-.77h-.095Zm.697-.688a.693.693 0 0 0-.697.688h.095c0-.328.27-.595.602-.595v-.093Zm.13 0h-.13v.093h.13v-.093Zm-.047-.725v.77h.095v-.77h-.095Zm.55-.673a.69.69 0 0 0-.55.673h.095c0-.285.203-.524.476-.582l-.02-.09Zm-.051-6.26.014 6.306h.095l-.014-6.306h-.095Zm-4.932-4.447c2.583 0 4.704 1.956 4.932 4.452l.094-.009c-.232-2.543-2.393-4.536-5.026-4.536v.093ZM5.932 10.84c.227-2.496 2.348-4.452 4.932-4.452v-.093c-2.633 0-4.795 1.993-5.027 4.536l.095.008Zm-.014 6.295.014-6.3h-.095l-.014 6.3h.095Zm.773.633c0-.18-.12-.337-.276-.453a1.236 1.236 0 0 0-.538-.226l-.014.093c.166.025.351.1.495.207.145.108.238.241.238.38h.095Zm0 .77v-.77h-.095v.77h.095Zm.082-.045h-.13v.093h.13v-.093Zm.697.688a.693.693 0 0 0-.697-.688v.093c.332 0 .602.267.602.595h.095Zm0 .77v-.77h-.095v.77h.095Zm-1.606.047h1.558v-.094H5.864v.094Zm-.108 0h.108v-.094h-.108v.094Zm-1.97 0h1.97v-.094h-1.97v.094Zm-3.636 0h3.636v-.094H.15v.094Zm-.048-.817v.77h.095v-.77H.102Zm.697-.688a.693.693 0 0 0-.697.688h.095c0-.328.27-.595.602-.595v-.093Zm.13 0h-.13v.093h.13v-.093Zm-.048-.725v.77h.095v-.77H.881Zm.698-.688a.693.693 0 0 0-.698.688h.095c0-.328.27-.594.603-.594v-.094ZM1.53 5.7v11.427h.095V5.698h-.095Zm-.667.046h.715v-.093H.864v.093Zm-.89-2.858L.82 5.712l.09-.027-.844-2.824-.09.026Zm3.812-.06H.02v.094h3.766v-.094ZM3.74.05v2.824h.095V.05h-.095Z' />
      </g>
      <defs>
        <clipPath id='a'>
          <path fill={fillColor[fill]} d='M0 0h22v20H0z' />
        </clipPath>
      </defs>
    </svg>
  )
}

const fillColor = {
  purple: '#7C65C1',
  white: '#FFFFFF',
}
