package corduroy;

import java.net.InetAddress;

/**
 * Node represents a single running instance where data can be stored and retrieved.
 */
public class Node {

    /**
     * Creates a node by passing in an address.
     * @param address An address in the format 'host:port'.
     */
    public Node(String address) {
        try {
            this.inetAddress = Utility.getInetAddress(address);
        }
        catch (Exception ex) {
            this.inetAddress = null;
        }
    }

    /**
     * Creates a node by passing in an {@link InetAddress} and an {@link int} that represents a Port.
     * @param inetAddress
     * @param port
     */
    public Node(InetAddress inetAddress, int port) {
        this.inetAddress = inetAddress;
        this.port = port;
    }

    /**
     * Gets the nodes address in the format 'host:port'.
     * @return The address.
     */
    public String getAddress() {
        return Utility.getAddress(inetAddress, port);
    }

    /**
     * Gets the hash value of the node with respect ot a specific ring number.
     * @param ring The number of the ring.
     * @return The hash value as a positive integer.
     */
    public int getHash(int ring) {
        try {
            return Utility.hash(getAddress(), ring);
        }
        catch (Exception ex) {
            return -1;
        }
    }

    /**
     * The {@link InetAddress} where the node is listening
     */
    private InetAddress inetAddress;

    /**
     * The {@link int} number of the Port where the node is listening.
     */
    private int port;
}
