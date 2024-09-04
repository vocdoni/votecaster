import { Box, Button, ButtonProps, Image } from '@chakra-ui/react'
import { FaUsers } from 'react-icons/fa6'
import { Link as RouterLink } from 'react-router-dom'
import { RoutePath } from '~constants'
import hat from '/degen-hat.png'

export const CreateFarcasterCommunityButton = () => (
  <RouterLink to={RoutePath.CommunitiesForm}>
    <Button size='sm' leftIcon={<FaUsers />}>
      Create your Farcaster community
    </Button>
  </RouterLink>
)

export const DegenButton = (props: ButtonProps) => (
  <Button display='flex' gap={2} fontWeight='500' {...props}>
    <Box width='1.2rem' height='1.2rem' lineHeight='1'>
      <Image src={hat} />
    </Box>{' '}
    {props.children}
  </Button>
)
