import { Box, Button, ButtonProps, Image } from '@chakra-ui/react'
import { Link } from 'react-router-dom'
import hat from '/degen-hat.png'

export const CreateFarcasterCommunityButton = () => (
  <Link to='/communities/new'>
    <DegenButton>Create your Farcaster community</DegenButton>
  </Link>
)

export const DegenButton = (props: ButtonProps) => (
  <Button display='flex' gap={2} fontWeight='500' {...props}>
    <Box width='1.2rem' height='1.2rem' lineHeight='1'>
      <Image src={hat} />
    </Box>{' '}
    {props.children}
  </Button>
)
