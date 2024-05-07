import {
  Avatar,
  Box,
  Code,
  Flex,
  FlexProps,
  Heading,
  IconButton,
  Image,
  Link,
  ListItem,
  Text,
  UnorderedList,
  useClipboard,
} from '@chakra-ui/react'
import { FaCheck, FaRegCopy } from 'react-icons/fa6'
import logo from '/full-logo.svg'

const devs = [
  {
    name: 'p4u',
    image:
      'https://res.cloudinary.com/merkle-manufactory/image/fetch/c_fill,f_png,w_168/https%3A%2F%2Fi.imgur.com%2FC877Bt1.png',
  },
  {
    name: 'elboletaire.eth',
    image:
      'https://res.cloudinary.com/merkle-manufactory/image/fetch/c_fill,f_jpg,w_168/https%3A%2F%2Fi.imgur.com%2FqLVpR8r.jpg',
  },
  {
    name: 'do',
    image:
      'https://res.cloudinary.com/merkle-manufactory/image/fetch/c_fill,f_jpg,w_168/https%3A%2F%2Fi.imgur.com%2F89SvQ0Q.jpg',
  },
  {
    name: 'kacuatro',
    image:
      'https://res.cloudinary.com/merkle-manufactory/image/fetch/c_fill,f_jpg,w_168/https%3A%2F%2Fi.imgur.com%2FA5TjFNp.jpg',
  },
  {
    name: 'ferran',
    image:
      'https://res.cloudinary.com/merkle-manufactory/image/fetch/c_fill,f_png,w_168/https%3A%2F%2Fipfs.decentralized-content.com%2Fipfs%2Fbafybeibgky4rmd6jdnczkmhg6mytakdxhugnueva7czt32rr2ry3hzasnq',
  },
]

export const Credits = (props: FlexProps) => {
  const { hasCopied, onCopy } = useClipboard('0x988A5a452D40aEB67B405eC7Dda6E28fe789646d')

  return (
    <Flex {...props}>
      <Flex gap={4} flexDir='column' maxW={600}>
        <Heading as='h1' size='md' textAlign='center'>
          Why Farcaster.vote?
        </Heading>
        <Text>
          <strong>Farcaster.vote</strong> introduces{' '}
          <strong>the first verifiable & decentralized polls within Farcaster Frames</strong>!
        </Text>
        <Text>
          As Farcaster grows, the demand for collective coordination solutions increases. However,{' '}
          <strong>centralized polls are not the best for decision-making</strong>, as the votes can't be verifiable,
          leaving room for vote tampering and censorship.
        </Text>
        <Text>
          This is where Farcaster.vote comes to play, <strong>combining Farcaster</strong>'s social network and identity
          system{` `}
          <strong>with the Vocdoni Protocol for tamper-proof and censorship-resistant digital voting</strong>,
          positioning Farcaster as the go-to platform for digital communities!
        </Text>
        {/* <Text>Read to know more ↗️</Text> */}
        <Flex flexDir='column' gap={8} align='center'>
          <Box>
            <Heading as='h2' size='sm' mb={4} textAlign='center'>
              Roadmap
            </Heading>
            <UnorderedList>
              <ListItem>Better UX design</ListItem>
              <ListItem>Delegated voting</ListItem>
              <ListItem>Multiple token strategy</ListItem>
              <ListItem>More rankings and statistics</ListItem>
              <ListItem>Gitcoin passport integration</ListItem>
              <ListItem>POAP event gated</ListItem>
              <ListItem>Voter rewards using $DEGEN for communities</ListItem>
              <ListItem>On-chain voting using frame verification on $DEGEN chain</ListItem>
              <ListItem>... and more!</ListItem>
            </UnorderedList>
          </Box>
          <Text fontWeight='bold' textAlign='center'>
            Do you want to create more flexible Web3 votes? <br />
            Check our Web3 voting UI for Ethereum based DAOs 👇
          </Text>
          <Box display='flex'>
            <Link href='https://onvote.app' target='_blank'>
              <Image src={logo} alt='onvote.app' maxW='150px' />
            </Link>
          </Box>
          <Box textAlign='center'>
            <Heading as='h3' size='md' mb={5}>
              Built with ❤️ by
            </Heading>
            <Flex direction='row' gap={3} justifyContent='space-between' px={10} wrap='wrap'>
              {devs.map((dev) => (
                <Box key={dev.name}>
                  <Link fontWeight='600' href={`https://warpcast.com/${dev.name}`} target='_blank'>
                    <Avatar name={dev.name} src={dev.image} />
                    <Text display='block'>@{dev.name}</Text>
                  </Link>
                </Box>
              ))}
            </Flex>
          </Box>
          <Box my={5} alignItems='center'>
            <Text float='left'>Tip us:</Text>
            <Flex>
              <Code
                as='span'
                fontStyle='italic'
                display='inline-block'
                ml={2}
                isTruncated
                maxW={{ base: 250, md: 400 }}
              >
                0x988A5a452D40aEB67B405eC7Dda6E28fe789646d
              </Code>
              <IconButton
                aria-label='Copy to clipboard'
                colorScheme='purple'
                icon={hasCopied ? <FaCheck /> : <FaRegCopy />}
                size='xs'
                onClick={onCopy}
                cursor='pointer'
                p={1.5}
                title={hasCopied ? 'Copied!' : 'Copy to clipboard'}
              />
            </Flex>
          </Box>
        </Flex>
      </Flex>
    </Flex>
  )
}
