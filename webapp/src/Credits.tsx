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
          Farcaster.vote is the first onchain voting system for Farcaster that enables voting securely within a Frame!
        </Text>
        <Text>
          As Farcaster expands its user base, there is a rising need for solutions for social coordination. But
          centralized polling fall short as the votes can't be verifiable, leaving room for vote tampering and
          censorship.
        </Text>
        <Text>
          This is where Farcaster.vote comes to play merging Farcaster technologies like Frames and FIDs with Vocdoni, a
          decentralized protocol for digital voting.
        </Text>
        <Text>
          Thanks to the magic of the Vocdoni protocol, we ensure that your votes are not only secure but also fully
          verifiable.
        </Text>
        {/* <Text>Read to know more ↗️</Text> */}
        <Flex flexDir='column' gap={8} align='center'>
          <Box>
            <Heading as='h2' size='sm' mb={4} textAlign='center'>
              Roadmap
            </Heading>
            <UnorderedList>
              <ListItem>Token-gated polls</ListItem>
              <ListItem>Channel-based polls</ListItem>
              <ListItem>Gitcoin passport gated</ListItem>
              <ListItem>POAP event gated</ListItem>
              <ListItem>Multiple token strategy polls</ListItem>
            </UnorderedList>
          </Box>
          <Text fontWeight='bold' textAlign='center'>
            Do you want to create more flexible Web3 votes? <br />
            Check our Web3 voting UI 👇
          </Text>
          <Box align='center'>
            <Link href='https://onvote.app' target='_blank'>
              <Image src={logo} alt='onvote.app' maxW='50%' />
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
