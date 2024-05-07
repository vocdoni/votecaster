import { Box, BoxProps, Heading, SimpleGrid, Text, VStack } from '@chakra-ui/react'

type FeatureProps = {
  title: string
  description: string | string[]
}

const Feature: React.FC<FeatureProps> = ({ title, description }) => (
  <Box p={5} boxShadow='md' bg='white' borderRadius='md'>
    <Heading size='md' mb={2} fontWeight={600}>
      {title}
    </Heading>
    <VStack color='gray.500' spacing={4} alignItems='start'>
      {description instanceof Array ? description.map((text) => <Text>{text}</Text>) : <Text>{description}</Text>}
    </VStack>
  </Box>
)

export const Features = (props: BoxProps) => (
  <Box p={5} {...props}>
    <Heading textAlign='center' mb={8}>
      Features for communities
    </Heading>
    <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} spacing={5}>
      <Feature
        title='Custom Censuses'
        description={[
          'Create a census for your community based on NFTs, ERC20 tokens or Farcaster channels.',
          'Ensure voting rights are exclusive to eligible members.',
        ]}
      />
      <Feature
        title='Notify voters'
        description='Boost poll participation by automatically notifying all community members directly on Farcaster every time there is a new poll!'
      />
      <Feature
        title='Framed polls'
        description='All polls are Frames. This means eligible members can cast their votes directly from the social feed in their preferred Farcaster client.'
      />
      <Feature
        title='Community Hub'
        description='A public site for your community that not only lists all your polls but also provides invaluable governance insights like voter turnout, active and non-active member lists, and many other information.'
      />
      <Feature
        title='Onchain communities and results'
        description='Deployed on Degenchain, your community benefits from immutable and transparent information. Ended poll results are also recorded on Degenchain for full accountability.'
      />
      <Feature
        title='E2E verifiable polls'
        description={[
          'Farcaster.vote leverages the decentralized Vocdoni Protocol to ensure that all voting is end-to-end verifiable, transparent, and secure!',
          'Coming soon: Onchain voting on Degenchain.',
        ]}
      />
    </SimpleGrid>
  </Box>
)
