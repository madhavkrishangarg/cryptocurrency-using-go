import subprocess
import time
import os
import json

def run_command(command, node_id):
    env = os.environ.copy()
    env['NODE_ID'] = str(node_id)
    result = subprocess.run(command, shell=True, capture_output=True, text=True, env=env)
    print(f"NODE {node_id}: {command}")
    print(result.stdout)
    if result.stderr:
        print(f"Error: {result.stderr}")
    print()
    return result.stdout

def create_wallet(node_id):
    output = run_command(f"./blockchain_go createwallet", node_id)
    address = output.strip().split(": ")[-1]
    return address

def main():
    central_node_id = 3000
    wallet_node_id = 3001
    miner_node_id = 3002

    central_node_address = create_wallet(central_node_id)
    wallet_address = create_wallet(wallet_node_id)
    miner_address = create_wallet(miner_node_id)

    wallet_addresses = {
        "CENTRAL_NODE": central_node_address,
        "WALLET": wallet_address,
        "MINER": miner_address,
    }
    with open("wallet_addresses.json", "w") as f:
        json.dump(wallet_addresses, f, indent=2)

    run_command(f"./blockchain_go createblockchain -address {central_node_address}", central_node_id)

    run_command(f"cp blockchain_{central_node_id}.db blockchain_genesis.db", central_node_id)

    print("Scenario setup complete. Please follow these steps in separate terminals:")
    print(f"1. In terminal 1, run: export NODE_ID={central_node_id} && ./blockchain_go startnode")
    print(f"2. In terminal 2, run: export NODE_ID={wallet_node_id} && cp blockchain_genesis.db blockchain_{wallet_node_id}.db && ./blockchain_go startnode")
    print(f"3. In terminal 3, run: export NODE_ID={miner_node_id} && cp blockchain_genesis.db blockchain_{miner_node_id}.db && ./blockchain_go startnode -miner {miner_address}")
    print("\nAfter starting all nodes, press Enter to continue with the transaction...")
    input()

    run_command(f"./blockchain_go send -from {central_node_address} -to {wallet_address} -amount 10 -mine", central_node_id)
    print("Sent initial coins. Wait a moment for it to be mined, then press Enter...")
    input()

    run_command(f"./blockchain_go send -from {wallet_address} -to {miner_address} -amount 1", wallet_node_id)
    print("Created transaction. Wait a moment for it to be mined, then press Enter...")
    input()

    run_command(f"./blockchain_go getbalance -address {wallet_address}", wallet_node_id)
    run_command(f"./blockchain_go getbalance -address {miner_address}", wallet_node_id)

    print("\nScenario complete. You can now interact with the nodes in their respective terminals.")
    print("Remember to stop the nodes when you're done.")

if __name__ == "__main__":
    main()
